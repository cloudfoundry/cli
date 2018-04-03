package crypto

import (
	"errors"
	"fmt"
	"io"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"os"
	"unicode"
)

type MultipleDigest struct {
	digests []Digest
}

func MustNewMultipleDigest(digests ...Digest) MultipleDigest {
	if len(digests) == 0 {
		panic("no digests have been provided")
	}
	return MultipleDigest{digests}
}

func MustParseMultipleDigest(json string) MultipleDigest {
	digest, err := ParseMultipleDigest(json)
	if err != nil {
		panic(fmt.Sprintf("Parsing multiple digest: %s", err))
	}
	return digest
}

func ParseMultipleDigest(json string) (MultipleDigest, error) {
	var digest MultipleDigest
	err := (&digest).UnmarshalJSON([]byte(json))
	if err != nil {
		return MultipleDigest{}, err
	}
	return digest, nil
}

func NewMultipleDigestFromPath(filePath string, fs boshsys.FileSystem, algos []Algorithm) (MultipleDigest, error) {
	file, err := fs.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return MultipleDigest{}, bosherr.WrapErrorf(err, "Calculating digest of '%s'", filePath)
	}
	defer func() {
		_ = file.Close()
	}()

	return NewMultipleDigest(file, algos)
}

func NewMultipleDigest(stream io.ReadSeeker, algos []Algorithm) (MultipleDigest, error) {
	if len(algos) == 0 {
		return MultipleDigest{}, errors.New("must provide at least one algorithm")
	}

	digests := []Digest{}
	for _, algo := range algos {
		stream.Seek(0, 0)
		digest, err := algo.CreateDigest(stream)
		if err != nil {
			return MultipleDigest{}, err
		}
		digests = append(digests, digest)

	}
	return MultipleDigest{digests}, nil
}

func (m MultipleDigest) Algorithm() Algorithm { return m.strongestDigest().Algorithm() }

func (m MultipleDigest) String() string {
	var result []string

	for _, digest := range m.digests {
		result = append(result, digest.String())
	}

	return strings.Join(result, ";")
}

func (m MultipleDigest) Verify(reader io.Reader) error {
	err := m.validate()
	if err != nil {
		return err
	}

	return m.strongestDigest().Verify(reader)
}

func (m MultipleDigest) VerifyFilePath(filePath string, fs boshsys.FileSystem) error {
	file, err := fs.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Calculating digest of '%s'", filePath)
	}
	defer func() {
		_ = file.Close()
	}()
	return m.Verify(file)
}

func (m MultipleDigest) validate() error {
	if len(m.digests) == 0 {
		return errors.New("Expected to find at least one digest")
	}

	algosUsed := map[string]struct{}{}

	for _, digest := range m.digests {
		algoName := digest.Algorithm().Name()

		if _, found := algosUsed[algoName]; found {
			return bosherr.Errorf("Multiple digests of the same algorithm '%s' found in digests '%s'", algoName, m.String())
		}

		algosUsed[algoName] = struct{}{}
	}

	return nil
}

func (m MultipleDigest) strongestDigest() Digest {
	if len(m.digests) == 0 {
		panic("no digests have been provided")
	}

	preferredAlgorithms := []Algorithm{DigestAlgorithmSHA512, DigestAlgorithmSHA256, DigestAlgorithmSHA1}

	for _, algo := range preferredAlgorithms {
		for _, digest := range m.digests {
			if digest.Algorithm().Name() == algo.Name() {
				return digest
			}
		}
	}

	return m.digests[0]
}

func (m *MultipleDigest) DigestFor(algo Algorithm) (Digest, error) {
	for _, digest := range m.digests {
		algoName := digest.Algorithm().Name()
		if algoName == algo.Name() {
			return digest, nil
		}
	}

	return nil, errors.New("digest-for-algorithm-not-present")
}

func (m *MultipleDigest) UnmarshalJSON(data []byte) error {
	digestString := strings.TrimSuffix(strings.TrimPrefix(string(data), `"`), `"`)

	multiDigest, err := m.parseMultipleDigestString(digestString)
	if err != nil {
		return err
	}

	err = multiDigest.validate()
	if err != nil {
		return err
	}

	*m = multiDigest

	return nil
}

func (m MultipleDigest) MarshalJSON() ([]byte, error) {
	if len(m.digests) == 0 {
		return nil, errors.New("no digests have been provided")
	}

	return []byte(fmt.Sprintf(`"%s"`, m.String())), nil
}

func (m MultipleDigest) parseMultipleDigestString(multipleDigest string) (MultipleDigest, error) {
	pieces := strings.Split(multipleDigest, ";")

	digests := []Digest{}

	for _, digest := range pieces {
		parsedDigest, err := m.parseDigestString(digest)
		if err == nil {
			digests = append(digests, parsedDigest)
		} else if _, ok := err.(emptyDigestError); !ok {
			return MultipleDigest{}, err
		}
	}

	if len(digests) == 0 {
		return MultipleDigest{}, errors.New("No digest algorithm found. Supported algorithms: sha1, sha256, sha512")
	}

	return MultipleDigest{digests: digests}, nil
}

type emptyDigestError struct{}

func (e emptyDigestError) Error() string {
	return "Empty digest parsed from digest string"
}

func (MultipleDigest) parseDigestString(digest string) (Digest, error) {
	if len(digest) == 0 {
		return nil, emptyDigestError{}
	}

	pieces := strings.SplitN(digest, ":", 2)

	for _, piece := range pieces {
		if !isStringAlphanumeric(piece) {
			return nil, errors.New("Unable to parse digest string. Digest and algorithm key can only contain alpha-numeric characters.")
		}
	}

	if len(pieces) == 1 {
		// historically digests were only sha1 and did not include a prefix.
		// continue to support that behavior.
		pieces = []string{"sha1", pieces[0]}
	}

	switch pieces[0] {
	case "sha1":
		return NewDigest(DigestAlgorithmSHA1, pieces[1]), nil
	case "sha256":
		return NewDigest(DigestAlgorithmSHA256, pieces[1]), nil
	case "sha512":
		return NewDigest(DigestAlgorithmSHA512, pieces[1]), nil
	default:
		return NewDigest(NewUnknownAlgorithm(pieces[0]), pieces[1]), nil
	}
}

func isStringAlphanumeric(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, runes := range s {
		if !isAlphanumeric(runes) {
			return false
		}
	}
	return true
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || unicode.IsDigit(r)
}
