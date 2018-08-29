package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

func MakeBuildpackArchive(stackName string) string {
	archiveFile, err := ioutil.TempFile("", "buildpack-archive-file-")
	Expect(err).ToNot(HaveOccurred())
	err = archiveFile.Close()
	Expect(err).ToNot(HaveOccurred())
	err = os.RemoveAll(archiveFile.Name())
	Expect(err).ToNot(HaveOccurred())

	archiveName := archiveFile.Name() + ".zip"

	dir, err := ioutil.TempDir("", "buildpack-dir-")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	manifest := filepath.Join(dir, "manifest.yml")
	err = ioutil.WriteFile(manifest, []byte(fmt.Sprintf("stack: %s", stackName)), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = Zipit(dir, archiveName, "")
	Expect(err).ToNot(HaveOccurred())

	return archiveName
}

func BuildpackWithStack(f func(buildpackArchive string), stackName string) {
	buildpackZip := MakeBuildpackArchive(stackName)
	defer os.Remove(buildpackZip)

	f(buildpackZip)
}

func BuildpackWithoutStack(f func(buildpackArchive string)) {
	BuildpackWithStack(f, "")
}
