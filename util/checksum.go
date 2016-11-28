package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

//go:generate counterfeiter . Sha1Checksum

type Sha1Checksum interface {
	ComputeFileSha1() ([]byte, error)
	CheckSha1(string) bool
	SetFilePath(string)
}

type sha1Checksum struct {
	filepath string
}

func NewSha1Checksum(filepath string) Sha1Checksum {
	return &sha1Checksum{
		filepath: filepath,
	}
}

func (c *sha1Checksum) ComputeFileSha1() ([]byte, error) {
	hash := sha1.New()

	f, err := os.Open(c.filepath)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	if _, err := io.Copy(hash, f); err != nil {
		return []byte{}, err
	}

	return hash.Sum(nil), nil
}

func (c *sha1Checksum) CheckSha1(targetSha1 string) bool {
	sha1, err := c.ComputeFileSha1()
	if err != nil {
		return false
	}

	if fmt.Sprintf("%x", sha1) == targetSha1 {
		return true
	}
	return false
}

func (c *sha1Checksum) SetFilePath(filepath string) {
	c.filepath = filepath
}
