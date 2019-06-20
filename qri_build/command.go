package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type command struct {
	String string
	Tmpl   []interface{}
	Dir    string
	Env    map[string]string
}

// Run executes a command
func (c command) Run() error {
	return c.prepare(false).Run()
}

// RunStdout executes a command, returning whatever is printed to stdout
// as a string
func (c command) RunStdout() (res string, err error) {
	buf := &bytes.Buffer{}
	cmd := c.prepare(false)
	cmd.Stdout = buf
	if err = cmd.Run(); err != nil {
		return
	}
	res = buf.String()
	return
}

func (c command) SecretRunStdout() (res string, err error) {
	buf := &bytes.Buffer{}
	cmd := c.prepare(true)
	cmd.Stdout = buf
	if err = cmd.Run(); err != nil {
		return
	}
	res = buf.String()
	return
}

func (c command) prepare(quiet bool) *exec.Cmd {
	str := fmt.Sprintf(c.String, c.Tmpl...)
	args := strings.Split(str, " ")
	name := args[0]

	if !quiet {
		if c.Dir != "" {
			log.Debugf("$ pwd %s", c.Dir)
		}
		log.Infof("$ %s %s", name, strings.Join(args[1:], " "))
	}

	cmd := exec.Command(name, args[1:]...)
	cmd.Dir = c.Dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = envs(c.Env)
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

func removeAll(path string) error {
	log.Infof("remove: %s", path)
	return os.RemoveAll(path)
}

func move(oldpath, newpath string) error {
	log.Infof("move: %s -> %s", oldpath, newpath)

	if err := os.MkdirAll(filepath.Dir(newpath), 0777); err != nil {
		return fmt.Errorf("error making directories: %s", err)
	}
	return os.Rename(oldpath, newpath)
}

// npm does funny things to PATH, this gets the npm'd path
// http://2ality.com/2016/01/locally-installed-npm-executables.html
func npmDoPath(pwd string) (path string, err error) {
	npmBinPath, err := command{
		String: "npm bin",
		Dir:    pwd,
	}.SecretRunStdout()

	npmBinPath = strings.TrimSpace(npmBinPath)

	if err != nil {
		return
	}

	path = os.Getenv("PATH")
	return fmt.Sprintf("%s:%s", npmBinPath, path), nil
}
