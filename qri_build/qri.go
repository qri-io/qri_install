package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

const binName = "qri"

func buildDir(platform, arch string) string {
	return fmt.Sprintf("%s_%s_%s", binName, platform, arch)
}

// BuildQri runs a build of the qri using the specified operating
// system and architecture
func BuildQri(platform, arch, qriRepoPath string) (err error) {
	dirName := buildDir(platform, arch)
	path := filepath.Join("./", dirName)
	binPath := filepath.Join(path, binName)

	log.Infof("building %s", dirName)

	// cleanup if already exists
	if fi, err := os.Stat(path); !os.IsNotExist(err) && fi.IsDir() {
		if err = CleanupQriBuild(platform, arch); err != nil {
			return err
		}
	}

	if err = os.Mkdir(path, os.ModePerm); err != nil {
		return
	}

	build := cmd{
		name: "go",
		args: []string{"build"},
		env: map[string]string{
			"GOOS":   platform,
			"GOARCH": arch,
			// TODO (b5): need this while we're still off go modules
			"GOPATH":      os.Getenv("GOPATH"),
			"GO111MODULE": "off",
		},
		flags: map[string]string{
			"o": binPath,
		},
	}

	return build.Run()
}

// BuildQriZip creates a zip archive from a qri binary, expects BuildQri for
// matching platform & arch has already been called
func BuildQriZip(platform, arch, templatesPath string) (err error) {
	name := fmt.Sprintf("%s_%s_%s.zip", binName, platform, arch)
	dirName := buildDir(platform, arch)
	path := filepath.Join("./", dirName)
	binPath := filepath.Join(path, binName)

	log.Infof("compressing %s", name)
	f, err := os.Create(name)
	if err != nil {
		return
	}

	zw := zip.NewWriter(f)

	binw, err := zw.Create(binName)
	if err != nil {
		return
	}
	binf, err := os.Open(binPath)
	if err != nil {
		return
	}
	if _, err = io.Copy(binw, binf); err != nil {
		return
	}

	tmpl, err := template.ParseGlob(filepath.Join(templatesPath, "**"))
	if err != nil {
		return
	}
	readmew, err := zw.Create("readme.md")
	if err != nil {
		return
	}
	err = tmpl.Lookup("qri_readme.md").Execute(readmew, map[string]string{
		"Platform": platform,
		"Arch":     arch,
	})
	if err != nil {
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
