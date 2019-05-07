package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"code.cloudfoundry.org/ykk"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func VerifyAppPackageContentsV3(appName string, files ...string) {
	tmpZipFilepath, err := ioutil.TempFile("", "")
	defer os.Remove(tmpZipFilepath.Name())
	Expect(err).ToNot(HaveOccurred())

	downloadFirstAppPackage(appName, tmpZipFilepath.Name())
	Expect(err).ToNot(HaveOccurred())

	info, err := tmpZipFilepath.Stat()
	Expect(err).ToNot(HaveOccurred())
	reader, err := ykk.NewReader(tmpZipFilepath, info.Size())
	Expect(err).ToNot(HaveOccurred())

	seenFiles := []string{}
	for _, file := range reader.File {
		seenFiles = append(seenFiles, file.Name)
	}

	Expect(seenFiles).To(ConsistOf(files))
}

func getFirstAppPackageGuid(appName string) string {
	session := CF("v3-packages", appName)
	Eventually(session).Should(Exit(0))

	myRegexp, err := regexp.Compile(GUIDRegex)
	Expect(err).NotTo(HaveOccurred())
	matches := myRegexp.FindAll(session.Out.Contents(), -1)
	packageGUID := matches[3]

	return string(packageGUID)
}

func downloadFirstAppPackage(appName string, tmpZipFilepath string) {
	appGUID := getFirstAppPackageGuid(appName)
	session := CF("curl", fmt.Sprintf("/v3/packages/%s/download", appGUID), "--output", tmpZipFilepath)
	Eventually(session).Should(Exit(0))
}

func VerifyAppPackageContentsV2(appName string, files ...string) {
	tmpZipFilepath, err := ioutil.TempFile("", "")
	defer os.Remove(tmpZipFilepath.Name())
	Expect(err).ToNot(HaveOccurred())

	downloadFirstAppBits(appName, tmpZipFilepath.Name())
	Expect(err).ToNot(HaveOccurred())

	info, err := tmpZipFilepath.Stat()
	Expect(err).ToNot(HaveOccurred())
	reader, err := ykk.NewReader(tmpZipFilepath, info.Size())
	Expect(err).ToNot(HaveOccurred())

	seenFiles := []string{}
	for _, file := range reader.File {
		seenFiles = append(seenFiles, file.Name)
	}

	Expect(seenFiles).To(ConsistOf(files))
}

func downloadFirstAppBits(appName string, tmpZipFilepath string) {
	appGUID := AppGUID(appName)
	session := CF("curl", fmt.Sprintf("/v2/apps/%s/download", appGUID), "--output", tmpZipFilepath)
	Eventually(session).Should(Exit(0))
}
