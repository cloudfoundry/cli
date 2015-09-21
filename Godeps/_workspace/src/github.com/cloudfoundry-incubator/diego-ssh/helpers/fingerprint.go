package helpers

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/ssh"
)

const MD5_FINGERPRINT_LENGTH = 47
const SHA1_FINGERPRINT_LENGTH = 59

func MD5Fingerprint(key ssh.PublicKey) string {
	md5sum := md5.Sum(key.Marshal())
	return hex(md5sum[:])
}

func SHA1Fingerprint(key ssh.PublicKey) string {
	sha1sum := sha1.Sum(key.Marshal())
	return hex(sha1sum[:])
}

func hex(data []byte) string {
	var fingerprint string
	for i := 0; i < len(data); i++ {
		fingerprint = fmt.Sprintf("%s%0.2x", fingerprint, data[i])
		if i != len(data)-1 {
			fingerprint = fingerprint + ":"
		}
	}
	return fingerprint
}
