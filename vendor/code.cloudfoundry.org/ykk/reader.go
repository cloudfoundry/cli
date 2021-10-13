// Package ykk was designed to read zip files with additional prefixed bytes
// (aka Java jar files).
package ykk // import "code.cloudfoundry.org/ykk"

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
)

// Data is the expected input for NewReader. Satisfied by io.File and
// bytes.Reader.
type Data interface {
	io.ReadSeeker
	io.ReaderAt
}

// NewReader returns a zip for the given file path.
func NewReader(sourceZip Data, sourceSize int64) (*zip.Reader, error) {
	offset, err := FirstLocalFileHeaderSignature(sourceZip)
	if err != nil {
		return nil, err
	}

	_, err = sourceZip.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	readerAt := io.NewSectionReader(sourceZip, offset, sourceSize-offset)
	return zip.NewReader(readerAt, sourceSize)
}

// FirstLocalFileHeaderSignature returns the relative offset of the header of
// the first file entry.
func FirstLocalFileHeaderSignature(source io.Reader) (int64, error) {
	// zip file header signature, 0x04034b50, reversed due to little-endian byte
	// order
	firstByte := byte(0x50)
	restBytes := []byte{0x4b, 0x03, 0x04}
	count := int64(-1)
	foundAt := int64(-1)

	reader := bufio.NewReader(source)

	keepGoing := true
	for keepGoing {
		count++

		b, err := reader.ReadByte()
		if err != nil {
			keepGoing = false
			break
		}

		if b == firstByte {
			nextBytes, err := reader.Peek(3)
			if err != nil {
				keepGoing = false
			}
			if bytes.Compare(nextBytes, restBytes) == 0 {
				foundAt = count
				keepGoing = false
				break
			}
		}
	}

	return foundAt, nil
}
