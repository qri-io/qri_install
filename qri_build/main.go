package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	log                                   = logrus.New()
	arches, platforms                     string
	repoPath, frontendPath, templatesPath string
)

// strEnvFlags maps configuration values to flags and environment variables
// each variable defaults to "defaultVal", can be set by an environment variable
// specified by the entry keym, or via a command-line flag. command-line flags
// override all enviornment variables
var strEnvFlags = map[string]struct {
	val        *string
	flag       string
	defaultVal string
	usage      string
}{
	"QRI_BUILD_PLATFORM":  {&platforms, "os", "", "build platforms split with commas (darwin|windows|linux)"},
	"QRI_BUILD_ARCH":      {&arches, "arch", "", "build architectures split multiples with commas (386|amd64|arm)"},
	"QRI_BUILD_QRI":       {&repoPath, "qri", "", "path to qri repository"},
	"QRI_BUILD_FRONTEND":  {&frontendPath, "frontend", "", "path to qri frontend repository"},
	"QRI_BUILD_TEMPLATES": {&templatesPath, "templates", "templates", "path to build template files"},
}

func init() {

	// strEnvFlags["QRI_REPO_PATH"].defaultVal = filepath.Join(os.Getenv("GOPATH"), "github.com/qri-io/qri")

	// configure flag package
	for _, fl := range strEnvFlags {
		flag.StringVar(fl.val, fl.flag, fl.defaultVal, fl.usage)
	}
}

func main() {
	args := parseFlags()

	if len(args) == 0 {
		flag.PrintDefaults()
		return
	}

	log.Infof("\n\tbuild: %s\n\tarches: %s\n\tplatforms: %s\n\trepoPath: %s\n\tfrontendPath: %s\n\ttemplatesPath: %s\n", args[0], arches, platforms, repoPath, frontendPath, templatesPath)

	var wg sync.WaitGroup
	switch strings.TrimSpace(strings.ToLower(args[0])) {
	case "qri":
		for _, arch := range strings.Split(arches, ",") {
			for _, platform := range strings.Split(platforms, ",") {
				wg.Add(1)
				go func(arch, platform string) {
					if err := BuildQriZip(arch, platform, repoPath); err != nil {
						log.Errorf("%s", err.Error())
					}
					wg.Done()
				}(arch, platform)
			}
		}
	case "electron":
		log.Errorf("unfinished: %s", args[0])
	case "webapp":
		if err := BuildWebapp(frontendPath); err != nil {
			log.Errorf("building webapp: %s", err)
		}
	default:
		log.Errorf("unrecognized subcommand: %s", args[0])
	}

	wg.Wait()
}

func parseFlags() []string {
	flag.Parse()

	// check to see if flags are default and environment is set, overriding if so
	for key, def := range strEnvFlags {
		env := os.Getenv(key)
		if env != "" && *def.val == def.defaultVal {
			*def.val = env
		}
	}

	return flag.Args()
}

type command struct {
	String string
	Tmpl   []interface{}
	dir    string
	env    map[string]string
}

// Run executes a command
func (c command) Run() error {
	return c.prepare().Run()
}

// RunStdout executes a command, returning whatever is printed to stdout
// as a string
func (c command) RunStdout() (res string, err error) {
	buf := &bytes.Buffer{}
	cmd := c.prepare()
	cmd.Stdout = buf
	if err = cmd.Run(); err != nil {
		return
	}
	res = buf.String()
	return
}

func (c command) prepare() *exec.Cmd {
	str := fmt.Sprintf(c.String, c.Tmpl...)
	args := strings.Split(str, " ")
	name := args[0]
	log.Infof("%s %s", name, strings.Join(args[1:], " "))
	cmd := exec.Command(name, args[1:]...)
	cmd.Dir = c.dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = envs(c.env)
	return cmd
}

// RunCommands calls run on a series of commands
func RunCommands(cs ...command) (err error) {
	for _, cmd := range cs {
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("running %s: %s", cmd.String, err)
		}
	}
	return
}

// envs converts a map of var : value environment variables to a slice of
// key=value strings, suitable for os/exec.Cmd.Env
func envs(vars map[string]string) (envs []string) {
	for key, val := range vars {
		envs = append(envs, fmt.Sprintf("%s=%s", key, val))
	}
	return
}

func flags(vars map[string]string) (flags []string) {
	for key, val := range vars {
		if val == "true" {
			flags = append(flags, fmt.Sprintf("-%s", key))
		} else {
			flags = append(flags, fmt.Sprintf("-%s=%s", key, val))
		}
	}
	return
}
