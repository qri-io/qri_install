package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// BuildWebapp builds the frontend and moves the result into a local directory
// The core logic executes the following command from the frontendPath directory
// 	cross-env NODE_OPTIONS=\"--max_old_space_size=10000\" NODE_ENV=production
// 	node
// 		--trace-warnings
// 		-r @babel/register
// 		./node_modules/webpack/bin/webpack
// 			--config webpack.config.webapp.prod.js
// 			--colors
func BuildWebapp(frontendPath string) (err error) {

	// webpackPath := filepath.Join(frontendPath, "node_modules/webpack/bin/webpack")
	outputPath := filepath.Join(frontendPath, "dist/web")

	path, err := npmDoPath(frontendPath)
	if err != nil {
		return
	}

	cmd := command{
		String: "node --trace-warnings --require @babel/register %s --config %s --colors",
		Tmpl: []interface{}{
			"node_modules/webpack/bin/webpack",
			"webpack.config.webapp.prod.js",
		},
		dir: frontendPath,
		env: map[string]string{
			"PATH":         path,
			"NODE_ENV":     "production",
			"NODE_OPTIONS": "--max_old_space_size=10000",
		},
	}

	if err = cmd.Run(); err != nil {
		return err
	}

	return os.Rename(outputPath, "./web")
}

// http://2ality.com/2016/01/locally-installed-npm-executables.html
func npmDoPath(pwd string) (path string, err error) {
	npmBinPath, err := command{
		String: "npm bin",
		dir:    pwd,
	}.RunStdout()

	if err != nil {
		return
	}

	path = os.Getenv("PATH")
	return fmt.Sprintf("%s:%s", npmBinPath, path), nil
}
