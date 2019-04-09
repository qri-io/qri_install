package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ElectronCmd builds the qri electron app
var ElectronCmd = &cobra.Command{
	Use: "electron",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ElectronBuild("", ""); err != nil {
			log.Error("building electron: %s", err)
		}
	},
}

// ElectronBuild runs main and render processes
func ElectronBuild(platforms, arch string) (err error) {
	// cross-env NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.main.prod.js --colors
	return fmt.Errorf("not finished")
}

// ElectronBuildMain builds the main process (electron backend)
func ElectronBuildMain(platforms, arch string) (err error) {
	// "electron:build:main": "cross-env NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.main.prod.js --colors"
	return fmt.Errorf("not finished")
}

// ElectronBuildRenderer builds the render process (react frontend)
func ElectronBuildRenderer(platforms, arch string) (err error) {
	// "electron:build:renderer": "cross-env NODE_OPTIONS=\"--max_old_space_size=10000\" NODE_ENV=production node --trace-warnings -r @babel/register ./node_modules/webpack/bin/webpack --config webpack.config.renderer.prod.js --colors"
	return fmt.Errorf("not finished")
}
