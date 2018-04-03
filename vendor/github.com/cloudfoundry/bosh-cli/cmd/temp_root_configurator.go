package cmd

import (
	"github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type TempRootConfigurator interface {
	PrepareAndSetTempRoot(path string, logger logger.Logger) error
}

type tempRootConfigurator struct {
	fs boshsys.FileSystem
}

func NewTempRootConfigurator(fs boshsys.FileSystem) TempRootConfigurator {
	return &tempRootConfigurator{fs: fs}
}

func (c *tempRootConfigurator) PrepareAndSetTempRoot(path string, logger logger.Logger) error {
	logger.Info("tempRootConfigurator", "Preparing temp root: %s", path)

	if c.fs.FileExists(path) {
		logger.Info("tempRootConfigurator", "Path exists, deleting")
		err := c.fs.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	logger.Info("tempRootConfigurator", "Setting file system temp root")
	err := c.fs.ChangeTempRoot(path)
	if err != nil {
		return err
	}

	return nil
}
