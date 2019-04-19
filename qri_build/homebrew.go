package main

import (
	"crypto/sha256"
	"io/ioutil"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// HomebrewCmd builds the qri homebrew installer
var HomebrewCmd = &cobra.Command{
	Use:   "homebrew",
	Short: "build the qri homebrew installer",
	Run: func(cmd *cobra.Command, args []string) {
		srcPath, err := cmd.Flags().GetString("src")
		if err != nil {
			log.Error(err)
			return
		}

		zipFile, err := cmd.Flags().GetString("zip")
		if err != nil {
			log.Error(err)
			return
		}

		ignoreDevRestriction, err := cmd.Flags().GetBool("ignore-dev-restriction")
		if err != nil {
			log.Error(err)
			return
		}

		if err := HomebrewBuildInstaller(srcPath, zipFile, ignoreDevRestriction); err != nil {
			log.Errorf("building homebrew: %s", err)
		}
	},
}

func init() {
	HomebrewCmd.Flags().String("src", "", "path to qri source repository")
	HomebrewCmd.Flags().String("zip", "", "zip file for release to publish")
	HomebrewCmd.Flags().Bool("ignore-dev-restriction", false, "whether to ignore the error about dev versions")
}

const homebrewFormulaTemplate = `
class Qri < Formula
  desc "Global dataset version control system built on the distributed web"
  homepage "https://qri.io/"
  url "https://github.com/qri-io/qri/releases/download/$VERSION/$ZIPFILE"
  version "$VERSION"
  sha256 "$SHA256"

  def install
    bin.install "qri"
  end

  test do
    system "#{bin}/qri", "version"
  end
end
`

// HomebrewBuildInstaller builds the homebrew installer
func HomebrewBuildInstaller(srcPath, zipFile string, ignoreDevRestriction bool) error {
	if srcPath == "" {
		return fmt.Errorf("required flag: --src <path to qri source>")
	}
	if zipFile == "" {
		return fmt.Errorf("required flag: --zip <path to zip release>")
	}

	// Calculate the sha256 of the zip file that is being released.
	data, err := ioutil.ReadFile(zipFile)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	hashDigest := fmt.Sprintf("%x", sum)

	// Get filename for the zip file that is being released.
	zipBasename := path.Base(zipFile)

	// Read the sourcefile that contains the current version number.
	libSourcefile := filepath.Join(srcPath, "lib/lib.go")
	data, err = ioutil.ReadFile(libSourcefile)
	if err != nil {
		return err
	}
	codeText := string(data)
	// Parse the version number from the sourcefile.
	versionLine, err := grep(codeText, "const VersionNumber")
	if err != nil {
		return err
	}
	versionParts := strings.Split(versionLine, " ")
	versionNum := strings.Replace(versionParts[3], "\"", "", -1)

	// It is an error to publish a development version.
	if strings.Contains(versionNum, "-dev") && !ignoreDevRestriction {
		return fmt.Errorf("Cannot publish a development version to homebrew: \"%s\"", versionNum)
	}

	// Replace vars in the template.
	content := homebrewFormulaTemplate
	content = strings.Replace(content, "$VERSION", versionNum, -1)
	content = strings.Replace(content, "$ZIPFILE", zipBasename, -1)
	content = strings.Replace(content, "$SHA256", hashDigest, -1)

	// TODO: Publish this to github.com/qri-io/homebrew-qri/Formula/qri.rb
	fmt.Printf(content)

	return nil
}

func grep(haystack, needle string) (string, error) {
	lines := strings.Split(haystack, "\n")
	for _, ln := range lines {
		if strings.Contains(ln, needle) {
			return ln, nil
		}
	}
	return "", fmt.Errorf("not found")
}
