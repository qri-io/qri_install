package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
	log.Infof("$ %s %s", name, strings.Join(args[1:], " "))
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
