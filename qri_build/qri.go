package main

import (
	"archive/zip"
	"fmt"
	"time"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"text/template"

	"github.com/spf13/cobra"
)

const binName = "qri"

// QriCmd is the command for building the qri go binary
var QriCmd = &cobra.Command{
	Use:   "qri",
	Short: "build the qri go binary",
	Run: func(cmd *cobra.Command, args []string) {

		arches, err := cmd.Flags().GetStringSlice("arches")
		if err != nil {
			log.Error(err)
			return
		}

		platforms, err := cmd.Flags().GetStringSlice("platforms")
		if err != nil {
			log.Error(err)
			return
		}

		repoPath, err := cmd.Flags().GetString("qri")
		if err != nil {
			log.Error(err)
			return
		}

		templatesPath, err := cmd.Flags().GetString("templates")
		if err != nil {
			log.Error(err)
			return
		}

		log.Debugf("\n\tbuild qri zip.\n\tarches: %s\n\tplatforms: %s\n\trepoPath: %s\n\ttemplatesPath: %s\n", arches, platforms, repoPath, templatesPath)

		var wg sync.WaitGroup
		for _, arch := range arches {
			for _, platform := range platforms {
				wg.Add(1)
				go func(arch, platform string) {
					if err := BuildQriZip(platform, arch, repoPath, templatesPath); err != nil {
						log.Errorf("%s", err.Error())
					}
					wg.Done()
				}(arch, platform)
			}
		}
		wg.Wait()
	},
}

func init() {
	QriCmd.Flags().String("qri", "qri", "path to qri repository")
	QriCmd.Flags().StringSlice("platforms", []string{runtime.GOOS}, "platforms to compile (darwin|windows|linux|...)")
	QriCmd.Flags().StringSlice("arches", []string{runtime.GOARCH}, "architectures to compile (386|amd64|arm|...)")
	QriCmd.Flags().String("templates", "templates", "path to templates directory")
}

// BuildQriZip constructs a zip archive from a qri binary with a
// templated readme
func BuildQriZip(platform, arch, qriRepoPath, templatesPath string) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("getting current directory: %s", err)
		return
	}
	if _, err = BuildQri(platform, arch, qriRepoPath); err != nil {
		log.Errorf("building qri: %s", err)
		return
	}
	os.Chdir(cwd)
	if err = ZipQriBuild(platform, arch, templatesPath); err != nil {
		log.Errorf("writing qri zip: %s", err)
		return
	}
	os.Chdir(cwd)
	if err = CleanupQriBuild(platform, arch); err != nil {
		log.Errorf("cleanup: %s", err)
		return
	}
	log.Infof("built zip")
	return
}

func buildDir(platform, arch string) string {
	return fmt.Sprintf("%s_%s_%s", binName, platform, arch)
}

// BuildQri runs a build of the qri using the specified operating
// system and architecture
func BuildQri(platform, arch, qriRepoPath string) (path string, err error) {
	created := time.Now()
	dirName := buildDir(platform, arch)
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path = filepath.Join(cwd, dirName)
	binPath := filepath.Join(dirName, binName)
	relPath, err := filepath.Rel(qriRepoPath, cwd)
	if err != nil {
		panic(err)
	}
	relBinPath := filepath.Join(relPath, binPath)

	// cleanup if already exists
	if fi, err := os.Stat(path); !os.IsNotExist(err) && fi.IsDir() {
		if err = CleanupQriBuild(platform, arch); err != nil {
			return "", err
		}
	}

	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}

	// GOCACHE is a required env variable in order to build
	// need to get the user's cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}

	// With go modules enabled, `go build` requires being in the directory of the build target.
	log.Infof("changing directory to %s", qriRepoPath)
	os.Chdir(qriRepoPath)

	build := command{
		String: "go build -o %s",
		Tmpl: []interface{}{
			relBinPath,
		},
		Env: map[string]string{
			"GOOS":   platform,
			"GOARCH": arch,
			"PATH":   os.Getenv("PATH"),
			// TODO (b5): need this while we're still off go modules
			"GOPATH":      os.Getenv("GOPATH"),
			"GO111MODULE": os.Getenv("GO111MODULE"),
			"GOCACHE":     cacheDir,
		},
	}

	return path, build.Run()
}

// ZipQriBuild creates a zip archive from a qri binary, expects BuildQri for
// matching platform & arch has already been called
func ZipQriBuild(platform, arch, templatesPath string) (err error) {
	name := fmt.Sprintf("%s_%s_%s.zip", binName, platform, arch)
	dirName := buildDir(platform, arch)
	binPath := filepath.Join(dirName, binName)

	log.Infof("compressing %s. binPath: %s templatesPath: %s", name, binPath, templatesPath)
	f, err := os.Create(name)
	if err != nil {
		log.Errorf("creating zip archive: %s", err)
		return
	}

	zw := zip.NewWriter(f)

	binFileHeader := &zip.FileHeader{
		Name: binName,
		Modified: created,

		CreatorVersion: (3 << 8), // indicate a unix-style zip creator version
		ExternalAttrs: (0777 << 16), // set permisisons to 0777
	}

	binw, err := zw.CreateHeader(binFileHeader)
	if err != nil {
		log.Errorf("creating zip bin: %s", err)
		return
	}
	binf, err := os.Open(binPath)
	if err != nil {
		log.Errorf("opening binPath: %s", err)
		return
	}
	if _, err = io.Copy(binw, binf); err != nil {
		log.Errorf("copying bin to zip archive: %s", err)
		return
	}

	tmpl, err := template.ParseGlob(filepath.Join(templatesPath, "**"))
	if err != nil {
		log.Infof("parsing templates: %s", err)
		return
	}

	readmeHeader := &zip.FileHeader{
		Name: "readme.md",
		Modified: created,

		CreatorVersion: (3 << 8), // indicate a unix-style zip creator version
		ExternalAttrs: (0644 << 16), // set permisisons
	}
	readmew, err := zw.CreateHeader(readmeHeader)
	if err != nil {
		log.Infof("create readme file: %s", err.Error())
		return
	}
	err = tmpl.Lookup("qri_readme.md").Execute(readmew, map[string]string{
		"Platform": platform,
		"Arch":     arch,
	})
	if err != nil {
		log.Errorf("opening binPath: %s", err)
		return
	}

	return zw.Close()
}

// CleanupQriBuild removes the temp build directory
func CleanupQriBuild(platform, arch string) (err error) {
	dirName := buildDir(platform, arch)
	path := filepath.Join("./", dirName)

	return os.RemoveAll(path)
}
