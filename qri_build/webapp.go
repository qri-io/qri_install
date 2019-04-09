package main

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

// WebappCmd builds the qri react-js frontend
var WebappCmd = &cobra.Command{
	Use:   "webapp",
	Short: "build the qri frontend webapp",
	Run: func(cmd *cobra.Command, args []string) {
		frontendPath, err := cmd.Flags().GetString("frontend")
		if err != nil {
			log.Error(err)
			return
		}

		if err := BuildWebapp(frontendPath); err != nil {
			log.Errorf("building webapp: %s", err)
		}
	},
}

func init() {
	WebappCmd.Flags().String("frontend", "frontend", "path to qri frontend repo")
}

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
		Dir: frontendPath,
		Env: map[string]string{
			"PATH":         path,
			"NODE_ENV":     "production",
			"NODE_OPTIONS": "--max_old_space_size=10000",
		},
	}

	if err = cmd.Run(); err != nil {
		return err
	}

	return move(outputPath, "./web")
}
