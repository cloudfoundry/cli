package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
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

func BuildpacksOutputRegex(fields BuildpackFields) string {
	anyStringRegex := `\S+`
	optionalStringRegex := `\S*`
	anyBoolRegex := `(true|false)`
	anyIntRegex := `\d+`

	nameRegex := anyStringRegex
	if fields.Name != "" {
		nameRegex = regexp.QuoteMeta(fields.Name)
	}

	positionRegex := anyIntRegex
	if fields.Position != "" {
		positionRegex = regexp.QuoteMeta(fields.Position)
	}

	enabledRegex := anyBoolRegex
	if fields.Enabled != "" {
		enabledRegex = regexp.QuoteMeta(fields.Enabled)
	}

	lockedRegex := anyBoolRegex
	if fields.Locked != "" {
		lockedRegex = regexp.QuoteMeta(fields.Locked)
	}

	filenameRegex := anyStringRegex
	if fields.Filename != "" {
		filenameRegex = regexp.QuoteMeta(fields.Filename)
	}

	stackRegex := optionalStringRegex
	if fields.Stack != "" {
		stackRegex = regexp.QuoteMeta(fields.Stack)
	}

	return fmt.Sprintf(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`, nameRegex, positionRegex, enabledRegex,
		lockedRegex, filenameRegex, stackRegex)
}

type BuildpackFields struct {
	Name     string
	Position string
	Enabled  string
	Locked   string
	Filename string
	Stack    string
}

func BuildpackWithoutStack(f func(buildpackArchive string)) {
	BuildpackWithStack(f, "")
}

func SetupBuildpackWithStack(buildpackName, stack string) {
	BuildpackWithStack(func(buildpackPath string) {
		session := CF("create-buildpack", buildpackName, buildpackPath, "99")
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))
	}, stack)
}

func SetupBuildpackWithoutStack(buildpackName string) {
	SetupBuildpackWithStack(buildpackName, "")
}
