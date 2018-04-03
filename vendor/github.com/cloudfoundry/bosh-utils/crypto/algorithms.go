package crypto

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

var (
	DigestAlgorithmSHA1   Algorithm = algorithmSHAImpl{"sha1"}
	DigestAlgorithmSHA256 Algorithm = algorithmSHAImpl{"sha256"}
	DigestAlgorithmSHA512 Algorithm = algorithmSHAImpl{"sha512"}
)

type algorithmSHAImpl struct {
	name string
}

func (a algorithmSHAImpl) Name() string { return a.name }

func (a algorithmSHAImpl) CreateDigest(reader io.Reader) (Digest, error) {
	hash := a.hashFunc()

	_, err := io.Copy(hash, reader)
	if err != nil {
		return nil, bosherr.WrapError(err, "Copying file for digest calculation")
	}

	return NewDigest(a, fmt.Sprintf("%x", hash.Sum(nil))), nil
}

func (a algorithmSHAImpl) hashFunc() hash.Hash {
	switch a.name {
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	default:
		panic("Internal inconsistency")
	}
}

type unknownAlgorithmImpl struct {
	name string
}

func NewUnknownAlgorithm(name string) unknownAlgorithmImpl {
	return unknownAlgorithmImpl{name: name}
}

func (c unknownAlgorithmImpl) Name() string { return c.name }

func (c unknownAlgorithmImpl) CreateDigest(reader io.Reader) (Digest, error) {
	return nil, bosherr.Errorf("Unable to create digest of unknown algorithm '%s'", c.name)
}
