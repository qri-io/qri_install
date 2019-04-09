package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// ElectronCmd builds the qri electron app
var ElectronCmd = &cobra.Command{
	Use:   "electron",
	Short: "build the qri electron app",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ElectronBuild("", "", "", ""); err != nil {
			log.Errorf("building electron: %s", err)
		}
	},
}

func init() {
	ElectronCmd.Flags().String("qri", "qri", "path to qri repository")
	ElectronCmd.Flags().StringSlice("platforms", []string{runtime.GOOS}, "platforms to compile (darwin|windows|linux|...)")
	ElectronCmd.Flags().StringSlice("arches", []string{runtime.GOARCH}, "architectures to compile (386|amd64|arm|...)")
	ElectronCmd.Flags().String("frontend", "frontend", "path to qri frontend repo")
}

// ElectronBuild runs main and render processes
func ElectronBuild(platforms, arch, frontendPath, qriPath string) (err error) {
	// fetch/checkout/init qri repo if not present
	// build qri go binary for required arches
	// move built binaries into frontend directories

	// fetch/checkout/init frontend repo if not present
	// concurrently build main & renderer
	return fmt.Errorf("not finished")
}

// ElectronBuildMain builds the main process (electron backend)
func ElectronBuildMain(platforms, arch, frontendPath string) (err error) {
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
			"NODE_ENV": "production",
			"PATH":     path,
		},
	}
	if err = cmd.Run(); err != nil {
		return
	}
	return fmt.Errorf("not finished")
}

// ElectronBuildRenderer builds the render process (react frontend)
func ElectronBuildRenderer(platforms, arch, frontendPath string) (err error) {
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
			"NODE_ENV": "production",
			"PATH":     path,
		},
	}
	if err = cmd.Run(); err != nil {
		return
	}
	return fmt.Errorf("not finished")
}
