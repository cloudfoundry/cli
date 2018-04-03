package main

import (
	"fmt"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

type opts struct {
	VerifyMultiDigestCommand MultiDigestCommand  `command:"verify-multi-digest"`
	CreateMultiDigestCommand CreateDigestCommand `command:"create-multi-digest"`
	VersionFlag              func() error        `long:"version"`
}

func main() {
	o := opts{}
	o.VersionFlag = func() error {
		return &flags.Error{
			Type:    flags.ErrHelp,
			Message: fmt.Sprintf("version %s\n", VersionLabel),
		}
	}

	_, err := flags.Parse(&o)

	if typedErr, ok := err.(*flags.Error); ok {
		if typedErr.Type == flags.ErrHelp {
			err = nil
		}
	}

	if err != nil {
		os.Exit(1)
	}
}

type MultiDigestArgs struct {
	File   string
	Digest string
}

type MultiDigestCommand struct {
	Args MultiDigestArgs `positional-args:"yes"`
}

func (m MultiDigestCommand) Execute(args []string) error {
	multipleDigest := boshcrypto.MustParseMultipleDigest(m.Args.Digest)
	file, err := os.Open(m.Args.File)
	if err != nil {
		return err
	}
	return multipleDigest.Verify(file)
}

type CreateDigestArgs struct {
	Algorithms string
	File       string
}

type CreateDigestCommand struct {
	Args CreateDigestArgs `positional-args:"yes"`
}

func (c CreateDigestCommand) Execute(args []string) error {
	algorithmStrs := strings.Split(c.Args.Algorithms, ",")
	algos := []boshcrypto.Algorithm{}
	for _, algorithmStr := range algorithmStrs {
		switch algorithmStr {
		case "sha1":
			algos = append(algos, boshcrypto.DigestAlgorithmSHA1)
		case "sha256":
			algos = append(algos, boshcrypto.DigestAlgorithmSHA256)
		case "sha512":
			algos = append(algos, boshcrypto.DigestAlgorithmSHA512)
		default:
			return bosherr.Errorf("unknown algorithm '%s'", algorithmStr)
		}
	}

	fs := boshsys.NewOsFileSystem(boshlog.NewLogger(boshlog.LevelNone))
	multipleDigest, err := boshcrypto.NewMultipleDigestFromPath(c.Args.File, fs, algos)
	if err != nil {
		return err
	}
	fmt.Printf("%s", multipleDigest.String())
	return nil
}
