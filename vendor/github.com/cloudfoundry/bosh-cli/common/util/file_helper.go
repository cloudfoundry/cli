package util

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	gopath "path"
	"path/filepath"
	"strings"
)

func AbsolutifyPath(pathToManifest string, pathToFile string, fs boshsys.FileSystem) (string, error) {
	if strings.HasPrefix(pathToFile, "http") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "file:///") || strings.HasPrefix(pathToFile, "/") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "file://~") {
		return pathToFile, nil
	}

	if strings.HasPrefix(pathToFile, "~") {
		return fs.ExpandPath(pathToFile)
	}

	var absPath string

	if !strings.HasPrefix(pathToFile, "file://") {
		absPath = filepath.Join(filepath.Dir(pathToManifest), pathToFile)
	} else {
		pathToFile = strings.Replace(pathToFile, "file://", "", 1)
		absPath = gopath.Join(gopath.Dir(pathToManifest), pathToFile)
		absPath = "file://" + absPath
	}

	return absPath, nil
}
