package tarball

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Cache interface {
	Get(source Source) (path string, found bool)
	Path(source Source) (path string)
	Save(sourcePath string, source Source) error
}

type cache struct {
	basePath string
	fs       boshsys.FileSystem
	logger   boshlog.Logger
	logTag   string
}

func NewCache(basePath string, fs boshsys.FileSystem, logger boshlog.Logger) Cache {
	return &cache{
		basePath: basePath,
		fs:       fs,
		logger:   logger,
		logTag:   "tarballCache",
	}
}

func (c *cache) Get(source Source) (string, bool) {
	cachedPath := c.Path(source)
	if c.fs.FileExists(cachedPath) {
		c.logger.Debug(c.logTag, "Found cached tarball at: '%s'", cachedPath)
		return cachedPath, true
	}

	return "", false
}

func (c *cache) Save(sourcePath string, source Source) error {
	err := c.fs.MkdirAll(c.basePath, os.FileMode(0766))
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to create cache directory '%s'", c.basePath)
	}

	err = boshfu.NewFileMover(c.fs).Move(sourcePath, c.Path(source))
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to save tarball path '%s' in cache", sourcePath)
	}

	c.logger.Debug(c.logTag, "Saving tarball in cache at: '%s'", c.Path(source))
	return nil
}

func (c *cache) Path(source Source) string {
	urlSHA1 := sha1.Sum([]byte(source.GetURL()))
	filename := fmt.Sprintf("%x-%s", string(urlSHA1[:]), source.GetSHA1())
	return filepath.Join(c.basePath, filename)
}
