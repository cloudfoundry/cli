package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"code.cloudfoundry.org/ykk"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type Package struct {
	Checksum string
}

func (p *Package) UnmarshalJSON(data []byte) error {
	var ccPackage struct {
		Data struct {
			Checksum struct {
				Value string `json:"value"`
			} `json:"checksum"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &ccPackage); err != nil {
		return err
	}

	p.Checksum = ccPackage.Data.Checksum.Value

	return nil
}

func GetFirstAppPackageGuid(appName string) string {
	session := CF("v3-packages", appName)
	Eventually(session).Should(Exit(0))

	// myRegexp, err := regexp.Compile("([a-f0-9-])\\s+ready")
	myRegexp, err := regexp.Compile(GUIDRegex)
	Expect(err).NotTo(HaveOccurred())
	matches := myRegexp.FindAll(session.Out.Contents(), -1)
	packageGUID := matches[3]

	return string(packageGUID)
}

func DownloadFirstAppPackage(appName string, tmpZipFilepath string) {
	packageGUID := GetFirstAppPackageGuid(appName)
	session := CF("curl", fmt.Sprintf("/v3/packages/%s/download", packageGUID), "--output", tmpZipFilepath)
	Eventually(session).Should(Exit(0))
	return
}

func VerifyAppPackageContents(appName string, files ...string) {
	tmpZipFilepath, err := ioutil.TempFile("", "")
	defer os.Remove(tmpZipFilepath.Name())
	Expect(err).ToNot(HaveOccurred())

	DownloadFirstAppPackage(appName, tmpZipFilepath.Name())
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
