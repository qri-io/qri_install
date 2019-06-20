package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ElectronCmd builds the qri electron app
var ElectronCmd = &cobra.Command{
	Use:   "electron",
	Short: "build the qri electron app",
	Run: func(cmd *cobra.Command, args []string) {
		qriPath, err := cmd.Flags().GetString("qri")
		if err != nil {
			log.Error(err)
			return
		}

		frontendPath, err := cmd.Flags().GetString("frontend")
		if err != nil {
			log.Error(err)
			return
		}

		publish, err := cmd.Flags().GetBool("publish")
		if err != nil {
			log.Error(err)
			return
		}

		if err := ElectronBuildPackage(frontendPath, qriPath, nil, nil, publish); err != nil {
			log.Errorf("building electron: %s", err)
		}
	},
}

func init() {
	ElectronCmd.Flags().String("qri", "qri", "path to qri repository")
	ElectronCmd.Flags().String("frontend", "frontend", "path to qri frontend repo")
	ElectronCmd.Flags().Bool("publish", false, "publish draft release to github, requires special access")
	// TODO (b5) - these are hardcoded for now
	// ElectronCmd.Flags().StringSlice("platforms", []string{runtime.GOOS}, "platforms to compile (darwin|windows|linux)")
	// ElectronCmd.Flags().StringSlice("arches", []string{runtime.GOARCH}, "architectures to compile (386|amd64|arm|...)")
}

// ElectronBuildPackage builds electron app components and packages 'em up
func ElectronBuildPackage(frontendPath, qriPath string, platforms, arches []string, publish bool) (err error) {
	path, err := npmDoPath(frontendPath)
	if err != nil {
		return
	}

	publishString := "never"
	gh_token := ""

	if publish {
		gh_token = os.Getenv("GH_TOKEN")
		if gh_token == "" {
			log.Error("You want to publish this release to github, but the \"GH_TOKEN\" environment variable is not set. Check out this article on how to get a personal access token from github: https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line")
			return
		}
		publishString = "always"
	}

	if err = ElectronBuild(frontendPath, qriPath, platforms, arches); err != nil {
		return err
	}

	cmd := command{
		String: "node_modules/.bin/build --publish %s",
		Tmpl: []interface{}{
			publishString,
		},
		Dir: frontendPath,
		Env: map[string]string{
			"PATH":     path,
			"GH_TOKEN": gh_token,
		},
	}

	if err = cmd.Run(); err != nil {
		log.Errorf("running build: %s", err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	electronDir := filepath.Join(cwd, "electron")
	if fi, err := os.Stat(electronDir); !os.IsNotExist(err) && fi.IsDir() {
		if err = removeAll(electronDir); err != nil {
			return err
		}
	}
	return move(filepath.Join(frontendPath, "dist"), electronDir)
}

// ElectronBuild runs main and render processes
func ElectronBuild(frontendPath, qriPath string, platforms, arches []string) (err error) {
	platform := "darwin"
	arch := "amd64"

	// TODO (b5) - fetch/checkout/init qri repo if not present
	// build qri go binary for required arches
	buildDirPath, err := BuildQri(platform, arch, qriPath)
	if err != nil {
		log.Errorf("building qri: %s", err)
		return
	}

	// move built binaries into frontend directories
	platformResourcesDir := filepath.Join(frontendPath, "resources", electronPlatform(platform))
	if err = removeAll(platformResourcesDir); err != nil {
		log.Errorf("error removing old platform resources: %s", err)
		return
	}
	if err = move(buildDirPath+"/qri", platformResourcesDir+"/qri"); err != nil {
		log.Errorf("error moving new plaform resources to frontend: %s", err)
		return
	}

	// TODO (b5) - fetch/checkout/init frontend repo if not present
	// concurrently build main & renderer
	errs := make(chan error)

	go func() {
		errs <- ElectronBuildMain(frontendPath, platforms, arches)
	}()
	go func() {
		errs <- ElectronBuildRenderer(frontendPath)
	}()

	if err = <-errs; err != nil {
		return
	}

	return <-errs
}

func electronPlatform(platform string) (eplat string) {
	switch platform {
	case "darwin":
		return "mac"
	}
	return platform
}

// ElectronBuildMain builds the main process (electron backend)
func ElectronBuildMain(frontendPath string, platforms, arches []string) (err error) {
	// "electron:build:main": "cross-env NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.main.prod.js --colors"
	path, err := npmDoPath(frontendPath)
	if err != nil {
		return
	}

	// cross-env NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.main.prod.js --colors
	cmd := command{
		String: "node --trace-warnings -r @babel/register %s --config %s --colors",
		Tmpl: []interface{}{
			"node_modules/webpack/bin/webpack",
			"webpack.config.main.prod.js",
		},
		Dir: frontendPath,
		Env: map[string]string{
			"PATH":     path,
			"NODE_ENV": "production",
		},
	}

	return cmd.Run()
}

// ElectronBuildRenderer builds the render process (react frontend)
func ElectronBuildRenderer(frontendPath string) (err error) {
	path, err := npmDoPath(frontendPath)
	if err != nil {
		return
	}

	// "electron:build:renderer": "cross-env NODE_OPTIONS=\"--max_old_space_size=10000\" NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.renderer.prod.js --colors"
	cmd := command{
		String: "node --trace-warnings -r @babel/register %s --config %s --colors",
		Tmpl: []interface{}{
			"node_modules/webpack/bin/webpack",
			"webpack.config.renderer.prod.js",
		},
		Dir: frontendPath,
		Env: map[string]string{
			"PATH":         path,
			"NODE_ENV":     "production",
			"NODE_OPTIONS": "--max_old_space_size=10000",
		},
	}

	return cmd.Run()
}
