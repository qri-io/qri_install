package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

// RootCmd is the root command
var RootCmd = &cobra.Command{
	Use:   "qri_build",
	Short: "CLI for building qri deliverables",
}

func init() {
	RootCmd.AddCommand(
		QriCmd,
		WebappCmd,
		ElectronCmd,
		HomebrewCmd,
	)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
