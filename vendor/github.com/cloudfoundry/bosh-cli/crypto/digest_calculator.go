package crypto

import (
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"strings"
)

type DigestCalculator interface {
	Calculate(string) (string, error)
	CalculateString(string) string
}

type digestCalculator struct {
	fs         boshsys.FileSystem
	algorithms []boshcrypto.Algorithm
}

func NewDigestCalculator(fs boshsys.FileSystem, algos []boshcrypto.Algorithm) DigestCalculator {
	return digestCalculator{
		fs:         fs,
		algorithms: algos,
	}
}

func (c digestCalculator) Calculate(filePath string) (string, error) {
	digest, err := boshcrypto.NewMultipleDigestFromPath(filePath, c.fs, c.algorithms)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Calculating digest for %s", filePath)
	}
	return digest.String(), nil
}

func (c digestCalculator) CalculateString(data string) string {
	digest, err := boshcrypto.NewMultipleDigest(strings.NewReader(data), c.algorithms)
	if err != nil {
		panic("According to the docs sha1.Write will never return an error")
	}
	return digest.String()
}
