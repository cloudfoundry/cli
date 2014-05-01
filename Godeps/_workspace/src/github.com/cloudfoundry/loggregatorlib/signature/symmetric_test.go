package signature

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleEncryption(t *testing.T) {
	key := "aaaaaaaaaaaaaaaa"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	decrypted, err := Decrypt(key, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, decrypted, message)
	assert.NotEqual(t, encrypted, message)
}

func TestEncryptionWithAShortKey(t *testing.T) {
	key := "short key"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	decrypted, err := Decrypt(key, encrypted)
	assert.NoError(t, err)

	assert.Equal(t, decrypted, message)
	assert.NotEqual(t, encrypted, message)
}

func TestDecryptionWithWrongKey(t *testing.T) {
	key := "short key"
	message := []byte("Super secret message that no one should read")
	encrypted, err := Encrypt(key, message)
	assert.NoError(t, err)

	_, err = Decrypt("wrong key", encrypted)
	assert.Error(t, err)
}

func TestThatEncryptionIsNonDeterministic(t *testing.T) {
	key := "aaaaaaaaaaaaaaaa"
	message := []byte("Super secret message that no one should read")
	encrypted1, err := Encrypt(key, message)
	assert.NoError(t, err)

	encrypted2, err := Encrypt(key, message)
	assert.NoError(t, err)

	assert.NotEqual(t, encrypted1, encrypted2)
}

//Test so that we are able to test that Ruby encrypts/signs the same way
func TestDigest(t *testing.T) {
	assert.Equal(t, DigestBytes([]byte("some-key")), []byte{0x68, 0x2f, 0x66, 0x97, 0xfa, 0x93, 0xec, 0xa6, 0xc8, 0x1, 0xa2, 0x32, 0x51, 0x9a, 0x9, 0xe3, 0xfe, 0xc, 0x5c, 0x33, 0x94, 0x65, 0xee, 0x53, 0xc3, 0xf9, 0xed, 0xf9, 0x2f, 0xd0, 0x1f, 0x35})
}

//Test so that we are able to test that Ruby encrypts/signs the same way
func TestForKeyCreation(t *testing.T) {
	key := "12345"
	assert.Equal(t, getEncryptionKey(key), []byte{0x59, 0x94, 0x47, 0x1a, 0xbb, 0x1, 0x11, 0x2a, 0xfc, 0xc1, 0x81, 0x59, 0xf6, 0xcc, 0x74, 0xb4})
}
