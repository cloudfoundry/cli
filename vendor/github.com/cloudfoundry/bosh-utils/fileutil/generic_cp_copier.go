package fileutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const genericCpCopierLogTag = "genericCpCopier"

type genericCpCopier struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewGenericCpCopier(
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Copier {
	return genericCpCopier{fs: fs, logger: logger}
}

func (c genericCpCopier) FilteredCopyToTemp(dir string, filters []string) (string, error) {
	var filtersFilesToCopy []string
	var err error

	filters = c.convertDirectoriesToGlobs(dir, filters)

	filesToCopy := []string{}

	for _, filterPath := range filters {
		filtersFilesToCopy, err = doublestar.Glob(filterPath)
		if err != nil {
			return "", bosherr.WrapError(err, "Finding files matching filters")
		}

		for _, fileToCopy := range filtersFilesToCopy {
			filesToCopy = append(filesToCopy, strings.TrimPrefix(strings.TrimPrefix(fileToCopy, dir), "/"))
		}
	}

	return c.tryInTempDir(func(tempDir string) error {
		for _, relativePath := range filesToCopy {
			src := filepath.Join(dir, relativePath)
			dst := filepath.Join(tempDir, relativePath)

			fileInfo, err := os.Stat(src)
			if err != nil {
				return bosherr.WrapErrorf(err, "Getting file info for '%s'", src)
			}

			if !fileInfo.IsDir() {
				err = c.cp(src, dst, tempDir)
				if err != nil {
					c.CleanUp(tempDir)
					return err
				}
			}
		}

		err = os.Chmod(tempDir, os.FileMode(0755))
		if err != nil {
			bosherr.WrapError(err, "Fixing permissions on temp dir")
		}

		return nil
	})
}

func (c genericCpCopier) tryInTempDir(fn func(string) error) (string, error) {
	tempDir, err := c.fs.TempDir("bosh-platform-commands-cpCopier-FilteredCopyToTemp")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary directory")
	}

	err = fn(tempDir)
	if err != nil {
		c.CleanUp(tempDir)
		return "", err
	}

	return tempDir, nil
}

func (c genericCpCopier) CleanUp(tempDir string) {
	err := c.fs.RemoveAll(tempDir)
	if err != nil {
		c.logger.Error(genericCpCopierLogTag, "Failed to clean up temporary directory %s: %#v", tempDir, err)
	}
}

func (c genericCpCopier) convertDirectoriesToGlobs(dir string, filters []string) []string {
	convertedFilters := []string{}
	for _, filter := range filters {
		src := filepath.Join(dir, filter)
		fileInfo, err := os.Stat(src)
		if err == nil && fileInfo.IsDir() {
			convertedFilters = append(convertedFilters, filepath.Join(src, "**", "*"))
		} else {
			convertedFilters = append(convertedFilters, src)
		}
	}

	return convertedFilters
}

func (c genericCpCopier) cp(src, dst, tempDir string) error {
	containingDir := filepath.Dir(dst)
	err := c.fs.MkdirAll(containingDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Making destination directory '%s' for '%s'", containingDir, src)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening destination file for copy '%s'", dst)
	}
	defer out.Close()

	// Open the input file
	in, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening source file for copy '%s'", src)
	}
	defer in.Close()

	// Copy inFilename input outFilename output
	_, err = io.Copy(out, in)
	if err != nil {
		return bosherr.WrapErrorf(err, "Copying source '%s' to destination '%s'", src, dst)
	}

	sfi, err := os.Stat(src)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting source file stats for '%s'", src)
	}

	err = os.Chmod(dst, sfi.Mode())
	if err != nil {
		return bosherr.WrapErrorf(err, "Changing file permissions for destination '%s'", dst)
	}

	return nil
}
