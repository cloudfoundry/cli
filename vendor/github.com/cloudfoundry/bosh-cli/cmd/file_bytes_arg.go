package cmd

import (
	"io/ioutil"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FileBytesArg struct {
	FS boshsys.FileSystem

	Bytes []byte
}

func (a *FileBytesArg) UnmarshalFlag(data string) error {
	if len(data) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	if data == "-" {
		bs, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return bosherr.WrapErrorf(err, "Reading from stdin")
		}

		(*a).Bytes = bs

		return nil
	}

	absPath, err := a.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting absolute path '%s'", data)
	}

	bytes, err := a.FS.ReadFile(absPath)
	if err != nil {
		return err
	}

	(*a).Bytes = bytes

	return nil
}
