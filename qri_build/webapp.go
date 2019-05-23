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

		readOnly, err := cmd.Flags().GetBool("readonly")
		if err != nil {
			log.Error(err)
			return
		}

		ipfsAdd, err := cmd.Flags().GetBool("ipfs")
		if err != nil {
			log.Error(err)
			return
		}

		apiURL, err := cmd.Flags().GetString("api-url")
		if err != nil {
			log.Error(err)
			return
		}

		if err := BuildWebapp(frontendPath, readOnly, ipfsAdd, apiURL); err != nil {
			log.Errorf("building webapp: %s", err)
		}
	},
}

func init() {
	WebappCmd.Flags().String("frontend", "frontend", "path to qri frontend repo")
	WebappCmd.Flags().Bool("readonly", false, "build webapp in readonly mode")
	WebappCmd.Flags().Bool("ipfs", false, "add completed build to local IPFS repo")
	WebappCmd.Flags().String("api-url", "", "url the webapp should ping for the qri api")
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
func BuildWebapp(frontendPath string, readOnly, ipfsAdd bool, apiURL string) (err error) {

	// webpackPath := filepath.Join(frontendPath, "node_modules/webpack/bin/webpack")
	outputPath := filepath.Join(frontendPath, "dist/web")
	if readOnly {
		outputPath = filepath.Join(frontendPath, "dist/readonly")
	}

	path, err := npmDoPath(frontendPath)
	if err != nil {
		return
	}

	webpackFile := "webpack.config.webapp.prod.js"
	if readOnly {
		webpackFile = "webpack.config.readonly.prod.js"
	}

	cmd := command{
		String: "node --trace-warnings --require @babel/register %s --config %s --colors",
		Tmpl: []interface{}{
			"node_modules/webpack/bin/webpack",
			webpackFile,
		},
		Dir: frontendPath,
		Env: map[string]string{
			"PATH":                       path,
			"NODE_ENV":                   "production",
			"NODE_OPTIONS":               "--max_old_space_size=10000",
			"QRI_FRONTEND_BUILD_API_URL": apiURL,
		},
	}

	if err = cmd.Run(); err != nil {
		return err
	}

	if err = move(outputPath, "./webapp"); err != nil {
		return
	}

	if ipfsAdd {
		hash, err := IPFSAdd("webapp")
		if err != nil {
			return err
		}
		log.Infof("ipfs hash: %s", hash)
	}

	return
}
