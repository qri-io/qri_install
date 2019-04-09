package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	"QRI_BUILD_PLATFORM":      {&platforms, "os", "", "build platforms split with commas (darwin|windows|linux)"},
	"QRI_BUILD_ARCH":          {&arches, "arch", "", "build architectures split multiples with commas (386|amd64|arm)"},
	"QRI_BUILD_REPO":          {&repoPath, "qri-repo", "", "path to qri repository"},
	"QRI_BUILD_FRONTEND_REPO": {&frontendPath, "frontend-repo", "", "path to qri frontend repository"},
	"QRI_BUILD_TEMPLATES":     {&templatesPath, "templates", "templates", "path to build template files"},
}

func init() {

	strEnvFlags["QRI_REPO_PATH"].defaultVal = filepath.Join(os.Getenv("GOPATH"), "github.com/qri-io/qri")

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
					defer wg.Done()

					if err := BuildQri(platform, arch, repoPath); err != nil {
						log.Errorf("building qri: %s", err)
						return
					}
					if err := BuildQriZip(platform, arch, templatesPath); err != nil {
						log.Errorf("writing qri zip: %s", err)
						return
					}
					if err := CleanupQriBuild(platform, arch); err != nil {
						log.Errorf("cleanup: %s", err)
						return
					}
					log.Infof("built zip")
				}(arch, platform)
			}
		}
	case "electron":
		log.Errorf("unfinished: %s", args[0])
	case "webapp":
		log.Errorf("unfinished: %s", args[0])
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

type cmd struct {
	name  string
	flags map[string]string
	env   map[string]string
	args  []string
}

func (cmd cmd) Run() error {
	command := exec.Command(cmd.name, append(cmd.args, flags(cmd.flags)...)...)
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Env = envs(cmd.env)
	return command.Run()
}

func runCommands(cmds []cmd) (err error) {
	for _, cmd := range cmds {
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("running %s: %s", cmd.name, err)
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
		flags = append(flags, fmt.Sprintf("-%s=%s", key, val))
	}
	return
}
