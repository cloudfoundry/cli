package glob

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Dir interface {
	Glob(patterns ...string) (filePaths []string, err error)
}

type dir struct {
	path string
}

func NewDir(path string) (d Dir) {
	return dir{path: path}
}

func (d dir) Glob(patterns ...string) (filePaths []string, err error) {
	for _, pattern := range patterns {
		var newFiles []string

		newFiles, err = d.glob(pattern)
		if err != nil {
			err = wrapError(err, "Finding files matching pattern %s", pattern)
			return
		}

		filePaths = append(filePaths, newFiles...)
	}

	return
}

func (d dir) glob(pattern string) (files []string, err error) {
	// unify path so that pattern will match files on Windows
	unifiedPath := strings.TrimPrefix(filepath.ToSlash(d.path), filepath.VolumeName(d.path))
	globPattern := path.Join(unifiedPath, pattern)

	glob, err := CompileGlob(globPattern)
	if err != nil {
		err = wrapError(err, "Compiling glob for pattern %s", pattern)
		return
	}

	filepath.Walk(d.path, func(path string, info os.FileInfo, inErr error) (err error) {
		path = strings.TrimPrefix(filepath.ToSlash(path), filepath.VolumeName(path))

		if inErr != nil {
			err = inErr
			return
		}

		if glob.Match(path) {
			files = append(files, strings.Replace(path, unifiedPath+"/", "", 1))
		}

		return
	})

	//	Ruby Dir.glob will include *.log when looking for **/*.log
	//	Our glob implementation will not do it automatically
	if strings.Contains(pattern, "**/*") {
		var extraFiles []string

		updatedPattern := strings.Replace(pattern, "**/*", "*", 1)
		extraFiles, err = d.glob(updatedPattern)
		if err != nil {
			err = wrapError(err, "Recursing into pattern %s", updatedPattern)
			return
		}

		files = append(files, extraFiles...)
	}
	return
}

func wrapError(err error, msg string, args ...interface{}) (newErr error) {
	msg = fmt.Sprintf(msg, args...)
	return errors.New(fmt.Sprintf("%s: %s", msg, err.Error()))
}
