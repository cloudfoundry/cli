package manifest_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/pushaction/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	var (
		pathToManifest string
		manifest       string
	)

	JustBeforeEach(func() {
		tempFile, err := ioutil.TempFile("", "manifest-test-")
		Expect(err).ToNot(HaveOccurred())
		Expect(tempFile.Close()).ToNot(HaveOccurred())
		pathToManifest = tempFile.Name()

		err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
	})

	Describe("ReadAndMergeManifests", func() {
		var (
			apps       []Application
			executeErr error
		)

		JustBeforeEach(func() {
			apps, executeErr = ReadAndMergeManifests(pathToManifest)
		})

		BeforeEach(func() {
			manifest = `---
applications:
- name: "app-1"
  buildpack: "some-buildpack"
  command: "some-command"
  health-check-http-endpoint: "\\some-endpoint"
  health-check-type: "http"
  instances: 10
  disk_quota: 100M
  memory: 200M
  stack: "some-stack"
  timeout: 120
- name: "app-2"
  disk_quota: 1G
  memory: 2G
- name: "app-3"
`
		})

		It("reads the manifest file", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(apps).To(ConsistOf(
				Application{
					Name:                    "app-1",
					Buildpack:               "some-buildpack",
					Command:                 "some-command",
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances:               10,
					DiskQuota:               100,
					Memory:                  200,
					StackName:               "some-stack",
					Timeout:                 120,
				},
				Application{
					Name:      "app-2",
					DiskQuota: 1024,
					Memory:    2048,
				},
				Application{Name: "app-3"},
			))
		})
	})
})
