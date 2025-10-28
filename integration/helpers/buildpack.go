package helpers

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// MakeBuildpackArchive makes a simple buildpack zip for a given stack.
func MakeBuildpackArchive(stackName string) string {
	archiveFile, err := os.CreateTemp("", "buildpack-archive-file-")
	Expect(err).ToNot(HaveOccurred())
	err = archiveFile.Close()
	Expect(err).ToNot(HaveOccurred())
	err = os.RemoveAll(archiveFile.Name())
	Expect(err).ToNot(HaveOccurred())

	archiveName := archiveFile.Name() + ".zip"

	dir, err := os.MkdirTemp("", "buildpack-dir-")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	manifest := filepath.Join(dir, "manifest.yml")
	err = os.WriteFile(manifest, []byte(fmt.Sprintf("stack: %s", stackName)), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = Zipit(dir, archiveName, "")
	Expect(err).ToNot(HaveOccurred())

	return archiveName
}

// BuildpackWithStack makes a simple buildpack for the given stack (using
// MakeBuildpackArchive) and yields it to the given function, removing the zip
// at the end.
func BuildpackWithStack(f func(buildpackArchive string), stackName string) {
	buildpackZip := MakeBuildpackArchive(stackName)
	defer os.Remove(buildpackZip)

	f(buildpackZip)
}

// BuildpackWithoutStack makes a simple buildpack without a stack (using
// MakeBuildpackArchive) and yields it to the given function, removing the zip
// at the end.
func BuildpackWithoutStack(f func(buildpackArchive string)) {
	BuildpackWithStack(f, "")
}

// SetupBuildpackWithStack makes and uploads a buildpack for the given stack.
func SetupBuildpackWithStack(buildpackName, stack string) {
	BuildpackWithStack(func(buildpackPath string) {
		session := CF("create-buildpack", buildpackName, buildpackPath, "99")
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))
	}, stack)
}

// SetupBuildpackWithoutStack makes and uploads a buildpack without a stack.
func SetupBuildpackWithoutStack(buildpackName string) {
	SetupBuildpackWithStack(buildpackName, "")
}

// CNB makes a simple cnb (using
// MakeCNBArchive) and yields it to the given function, removing the cnb file
// at the end.
func CNB(f func(buildpackArchive string)) {
	cnbArchive := MakeCNBArchive()
	defer os.Remove(cnbArchive)

	f(cnbArchive)
}

// MakeCNBArchive makes a simple buildpack cnb file for a given stack.
func MakeCNBArchive() string {
	archiveDir, err := os.MkdirTemp("", "cnb-archive-file-")
	Expect(err).ToNot(HaveOccurred())

	archiveFile, err := os.Create(filepath.Join(archiveDir, "buildpack.cnb"))
	Expect(err).ToNot(HaveOccurred())
	defer archiveFile.Close()

	tarWriter := tar.NewWriter(archiveFile)
	defer tarWriter.Close()

	manifestBytes := []byte(`{"schemaVersion": 2,"mediaType": "application/vnd.oci.image.index.v1+json","manifests": []}`)
	err = tarWriter.WriteHeader(&tar.Header{
		Name: "index.json",
		Mode: 0666,
		Size: int64(len(manifestBytes)),
	})
	Expect(err).ToNot(HaveOccurred())

	_, err = tarWriter.Write(manifestBytes)
	Expect(err).ToNot(HaveOccurred())

	layoutBytes := []byte(`{"imageLayoutVersion": "1.0.0"}`)
	err = tarWriter.WriteHeader(&tar.Header{
		Name: "oci-layout",
		Mode: 0666,
		Size: int64(len(layoutBytes)),
	})
	Expect(err).ToNot(HaveOccurred())

	_, err = tarWriter.Write(layoutBytes)
	Expect(err).ToNot(HaveOccurred())

	err = tarWriter.Flush()
	Expect(err).ToNot(HaveOccurred())

	return archiveFile.Name()
}

// BuildpackFields represents a buildpack, displayed in the 'cf buildpacks'
// command.
type BuildpackFields struct {
	Position  string
	Name      string
	Enabled   string
	Locked    string
	Filename  string
	Stack     string
	Lifecycle string
}

// DeleteBuildpackIfOnOldCCAPI deletes the buildpack if the CC API targeted
// by the current test run is <= 2.80.0. Before this version, some entities
// would receive and invalid next_url in paginated requests. Since our test run
// now generally creates more than 50 buildpacks, we need to delete test buildpacks
// after use if we are targeting an older CC API.
// see https://github.com/cloudfoundry/capi-release/releases/tag/1.45.0
func DeleteBuildpackIfOnOldCCAPI(buildpackName string) {
	minVersion := "2.99.0"
	if !IsVersionMet(minVersion) {
		deleteSessions := CF("delete-buildpack", buildpackName, "-f")
		Eventually(deleteSessions).Should(Exit())
	}
}

type Buildpack struct {
	GUID  string `json:"guid"`
	Name  string `json:"name"`
	Stack string `json:"stack"`
}

type BuildpackList struct {
	Buildpacks []Buildpack `json:"resources"`
}

func BuildpackGUIDByNameAndStack(buildpackName string, stackName string) string {
	url := "/v3/buildpacks?names=" + buildpackName
	if stackName != "" {
		url += "&stacks=" + stackName
	}
	session := CF("curl", url)
	bytes := session.Wait().Out.Contents()

	buildpacks := BuildpackList{}
	err := json.Unmarshal(bytes, &buildpacks)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(buildpacks.Buildpacks)).To(BeNumerically(">", 0))
	if stackName != "" {
		return buildpacks.Buildpacks[0].GUID
	}
	for _, buildpack := range buildpacks.Buildpacks {
		if buildpack.Stack == "" {
			return buildpack.GUID
		}
	}
	return ""
}
