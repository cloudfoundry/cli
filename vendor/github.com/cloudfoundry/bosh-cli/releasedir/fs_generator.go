package releasedir

import (
	"fmt"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FSGenerator struct {
	dirPath string
	fs      boshsys.FileSystem
}

func NewFSGenerator(dirPath string, fs boshsys.FileSystem) FSGenerator {
	return FSGenerator{dirPath: dirPath, fs: fs}
}

func (g FSGenerator) GenerateJob(name string) error {
	jobDirPath := filepath.Join(g.dirPath, "jobs", name)

	if g.fs.FileExists(jobDirPath) {
		return bosherr.Errorf("Job '%s' at '%s' already exists", name, jobDirPath)
	}

	err := g.fs.MkdirAll(jobDirPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' dir", name)
	}

	err = g.fs.MkdirAll(filepath.Join(jobDirPath, "templates"), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' templates dir", name)
	}

	specTpl := fmt.Sprintf(`---
name: %s

templates: {}

packages: []

properties: {}
`, name)

	err = g.fs.WriteFileString(filepath.Join(jobDirPath, "spec"), specTpl)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' spec file", name)
	}

	err = g.fs.WriteFileString(filepath.Join(jobDirPath, "monit"), "")
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' monit file", name)
	}

	return nil
}

func (g FSGenerator) GeneratePackage(name string) error {
	pkgDirPath := filepath.Join(g.dirPath, "packages", name)

	if g.fs.FileExists(pkgDirPath) {
		return bosherr.Errorf("Package '%s' at '%s' already exists", name, pkgDirPath)
	}

	err := g.fs.MkdirAll(pkgDirPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' dir", name)
	}

	specTpl := fmt.Sprintf(`---
name: %s

dependencies: []

files: []
`, name)

	err = g.fs.WriteFileString(filepath.Join(pkgDirPath, "spec"), specTpl)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' spec file", name)
	}

	err = g.fs.WriteFileString(filepath.Join(pkgDirPath, "packaging"), "set -e\n")
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' packaging file", name)
	}

	return nil
}
