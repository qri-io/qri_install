package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// DesktopCmd builds the qri desktop app
var DesktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "build the qri desktop app",
	Long: `
build the qri desktop app, by first buliding the qri command-line binary and copying it
into the desktop's folder. This command is dependent upon having a correct $GOPATH,
having $GO111MODULE set to 'on', and having both 'go' and 'yarn' installed.

The directories for the 'qri' and 'desktop' source code need to be specified as command-line
arguments. For convenience, both will have their code pulled from the git origin, which
requires them to have the 'master' branch checked out.

The final installed that is built will have it's path displayed once this process completes
without any errors.
`,
	Run: func(cmd *cobra.Command, args []string) {
		qriPath, err := cmd.Flags().GetString("qri")
		if err != nil {
			log.Error(err)
			return
		}

		desktopPath, err := cmd.Flags().GetString("desktop")
		if err != nil {
			log.Error(err)
			return
		}

		if err := DesktopBuildPackage(desktopPath, qriPath, nil, nil); err != nil {
			log.Errorf("%s", err)
		}
	},
}

func init() {
	DesktopCmd.Flags().String("qri", "", "path to qri repository")
	DesktopCmd.Flags().String("desktop", "", "path to qri desktop repo")
}

// RequiredGoVersion is the required version of go needed to build qri
const RequiredGoVersion = "1.12"

// DesktopBuildPackage builds the desktop app with the necessary qri binary
func DesktopBuildPackage(desktopPath, qriPath string, platforms, arches []string) (err error) {
	if qriPath == "" || desktopPath == "" {
		return fmt.Errorf("Flags --qri and --desktop are both required")
	}

	// Ensure source directories exist
	if _, err := os.Stat(qriPath); os.IsNotExist(err) {
		return fmt.Errorf("Directory \"%s\" does not exist", qriPath)
	}
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		return fmt.Errorf("Directory \"%s\" does not exist", desktopPath)
	}

	// Ensure valid go version, go modules
	log.Infof("ensuring valid go version and go modules support...")
	err = ensureGoEnvVars()
	if err != nil {
		return err
	}

	// Update source code for qri binary
	log.Infof("updating source code for qri...")
	if err = updateSource(qriPath); err != nil {
		return err
	}

	// Update source code for desktop app
	log.Infof("updating source code for desktop...")
	if err = updateSource(desktopPath); err != nil {
		return err
	}

	// Build qri binary
	log.Infof("building qri binary...")
	builtPath, err := buildQriBinary(qriPath)
	if err != nil {
		return err
	}

	// Copy qri binary into desktop's backend/ folder
	log.Infof("copying qri binary into desktop...")
	// Work-around for Windows slashes, would be ignored by path.Base
	targetBinName := path.Base(strings.Replace(builtPath, "\\", "/", -1))
	if runtime.GOOS == "windows" {
		// In Windows, make sure the binary ends in ".exe". If not, add the extension when
		// copying it.
		if !strings.HasSuffix(builtPath, ".exe") {
			targetBinName += ".exe"
		}
	}
	backendBinary := filepath.Join(desktopPath, "backend/", targetBinName)
	err = CopyFile(builtPath, backendBinary)
	if err != nil {
		return err
	}

	// Set the backend binary as executable
	err = os.Chmod(backendBinary, 0755)
	if err != nil {
		return err
	}

	// Build desktop app installer
	log.Infof("building desktop app installer...")
	err = buildDesktopApp(desktopPath)

	// Find built installer
	release, err := discoverDesktopInstaller(desktopPath)
	if err != nil {
		return err
	}

	fmt.Printf("Release installer at: %s", release)
	return nil
}

// updateSource ensures that the "master" branch is checked out, then pulls from the origin
func updateSource(path string) error {
	branchName, err := getCurrentGitBranch(path)
	if err != nil {
		return err
	}
	if branchName != "master" {
		return fmt.Errorf("Please switch \"%s\" to branch master, branch %s currently checked out",
			path, branchName)
	}

	err = doGitPull(path)
	if err != nil {
		return err
	}

	return nil
}

// buildQriBinary will build the qri binary, returning the path of the built binary
func buildQriBinary(projectPath string) (string, error) {
	buildPath := filepath.Join(projectPath, "build")
	targetBinPath := filepath.Join(buildPath, "qri")

	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		err := os.Mkdir(buildPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	cmd := command{
		String: "go build -o build/qri",
		Dir:    projectPath,
	}

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return targetBinPath, nil
}

// buildDesktopApp will build the distributable electron installer for desktop
func buildDesktopApp(path string) error {
	cmd := command{
		String: "yarn",
		Dir:    path,
	}

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = command{
		String: "yarn dist",
		Dir:    path,
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func discoverDesktopInstaller(path string) (string, error) {
	path = filepath.Join(path, "release")
	finfos, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	found := []string{}
	for _, fi := range finfos {
		if strings.HasSuffix(fi.Name(), ".exe") {
			found = append(found, filepath.Join(path, fi.Name()))
		}
		if strings.HasSuffix(fi.Name(), ".dmg") {
			found = append(found, filepath.Join(path, fi.Name()))
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no built installer found")
	}
	if len(found) > 1 {
		return "", fmt.Errorf("found multiple installers: %s", strings.Join(found, ", "))
	}
	return found[0], nil
}

func ensureGoEnvVars() error {
	cmd := command{
		String: "go version",
	}

	output, err := cmd.RunStdout()
	if err != nil {
		return err
	}

	// TODO(dlong): Switch to regex, and compare against minimum version
	if !strings.Contains(output, RequiredGoVersion) {
		return fmt.Errorf("")
	}

	value := os.Getenv("GO111MODULE")
	if value != "on" {
		return fmt.Errorf("Error: must set envvar `GO111MODULE` to `on`")
	}

	return nil
}

// getCurrentGitBranch returns the currently checked out git branch
func getCurrentGitBranch(path string) (string, error) {
	cmd := command{
		String: "git branch",
		Dir:    path,
	}

	output, err := cmd.RunStdout()
	if err != nil {
		return "", err
	}

	var branchName string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			branchName = strings.TrimSpace(line[2:])
		}
	}
	return branchName, nil
}

// doGitPull runs git pull
func doGitPull(path string) error {
	cmd := command{
		String: "git pull",
		Dir:    path,
	}
	return cmd.Run()
}

// CopyFile copies a file from "from" to "to"
func CopyFile(from, to string) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}
