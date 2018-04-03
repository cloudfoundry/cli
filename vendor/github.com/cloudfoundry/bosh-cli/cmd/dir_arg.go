package cmd

import (
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DirOrCWDArg struct {
	Path string
}

func (a *DirOrCWDArg) UnmarshalFlag(data string) error {
	if len(data) == 0 || data == "." {
		path, err := os.Getwd()
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting working directory")
		}

		*a = DirOrCWDArg{Path: path}
	} else {
		*a = DirOrCWDArg{Path: data}
	}

	return nil
}
