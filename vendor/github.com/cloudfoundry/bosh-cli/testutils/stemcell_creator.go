package testutils

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

func CreateStemcell(stemcellSrcDir string, stemcellPath string) error {
	session, err := RunCommand("tar", "-zcf", stemcellPath, "-C", stemcellSrcDir, ".")
	if err != nil {
		return err
	}

	if session.ExitCode() != 0 {
		return bosherr.Errorf("Failed to create stemcell src:'%s' dest:'%s'", stemcellSrcDir, stemcellPath)
	}

	return nil
}
