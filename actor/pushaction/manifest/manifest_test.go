package manifest_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/types"

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
		// There are additional tests for this function in manifest_*OS*_test.go

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
  buildpack: default
  disk_quota: 1G
  instances: 0
  memory: 2G
  services:
  - service_1
  - service_2
- name: "app-3"
  env:
    env_1: 'foo'
    env_2: 182837403930483038
    env_3: true
    env_4: 1.00001
`
		})

		It("reads the manifest file", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(apps).To(ConsistOf(
				Application{
					Name: "app-1",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "some-buildpack",
					},
					Command:                 "some-command",
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances: types.NullInt{
						Value: 10,
						IsSet: true,
					},
					DiskQuota:          100,
					Memory:             200,
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				},
				Application{
					Name: "app-2",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "",
					},
					DiskQuota: 1024,
					Instances: types.NullInt{
						IsSet: true,
						Value: 0,
					},
					Memory:   2048,
					Services: []string{"service_1", "service_2"},
				},
				Application{
					Name: "app-3",
					EnvironmentVariables: map[string]string{
						"env_1": "foo",
						"env_2": "182837403930483038",
						"env_3": "true",
						"env_4": "1.00001",
					},
				},
			))
		})
	})
})
