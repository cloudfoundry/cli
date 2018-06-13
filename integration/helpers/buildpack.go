package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

func BuildpackWithStack(f func(buildpackArchive string), stackName string) {

	archiveFile, err := ioutil.TempFile("", "buildpack-archive-file-")
	Expect(err).ToNot(HaveOccurred())
	err = os.Remove(archiveFile.Name())
	Expect(err).ToNot(HaveOccurred())

	buildpackZip := archiveFile.Name() + ".zip"

	dir, err := ioutil.TempDir("", "buildpack-dir-")
	Expect(err).ToNot(HaveOccurred())

	defer os.Remove(buildpackZip)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "bin")
	err = os.MkdirAll(path, 0755)
	Expect(err).ToNot(HaveOccurred())

	compileFile := filepath.Join(path, "compile")
	err = ioutil.WriteFile(compileFile, []byte("some-content"), 0666)
	Expect(err).ToNot(HaveOccurred())

	manifest := filepath.Join(dir, "manifest.yml")
	err = ioutil.WriteFile(manifest, []byte(fmt.Sprintf("stack: %s", stackName)), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = Zipit(dir, buildpackZip, "")
	Expect(err).ToNot(HaveOccurred())

	f(buildpackZip)
}
