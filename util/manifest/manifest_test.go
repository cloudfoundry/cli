package manifest_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/types"
	. "code.cloudfoundry.org/cli/util/manifest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	var manifest string

	Describe("ReadAndMergeManifests", func() {
		var (
			pathToManifest string
			apps           []Application
			executeErr     error
		)

		JustBeforeEach(func() {
			apps, executeErr = ReadAndMergeManifests(pathToManifest)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
		})

		// There are additional tests for this function in manifest_*OS*_test.go

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
  routes:
  - route: foo.bar.com
  - route: baz.qux.com
  services:
  - service_1
  - service_2
- name: "app-3"
  env:
    env_1: 'foo'
    env_2: 182837403930483038
    env_3: true
    env_4: 1.00001
- name: "app-4"
  buildpack: null
  command: null
`
			tempFile, err := ioutil.TempFile("", "manifest-test-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()

			err = ioutil.WriteFile(pathToManifest, []byte(manifest), 0666)
			Expect(err).ToNot(HaveOccurred())

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
					Command: types.FilteredString{
						IsSet: true,
						Value: "some-command",
					},
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances: types.NullInt{
						Value: 10,
						IsSet: true,
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 100,
						IsSet: true,
					},
					Memory: types.NullByteSizeInMb{
						Value: 200,
						IsSet: true,
					},
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				},
				Application{
					Name: "app-2",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "",
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 1024,
						IsSet: true,
					},
					Instances: types.NullInt{
						IsSet: true,
						Value: 0,
					},
					Memory: types.NullByteSizeInMb{
						Value: 2048,
						IsSet: true,
					},
					Routes:   []string{"foo.bar.com", "baz.qux.com"},
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
				Application{
					Name: "app-4",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "",
					},
					Command: types.FilteredString{
						IsSet: true,
						Value: "",
					},
				},
			))
		})
	})

	Describe("WriteApplicationManifest", func() {
		var (
			application Application
			tmpDir      string
			filePath    string

			executeErr error
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "manifest-test-")
			Expect(err).NotTo(HaveOccurred())
			filePath = filepath.Join(tmpDir, "manifest.yml")
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		JustBeforeEach(func() {
			executeErr = WriteApplicationManifest(application, filePath)
		})

		Context("when all app properties are provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "some-buildpack",
					},
					Command: types.FilteredString{
						IsSet: true,
						Value: "some-command",
					},
					EnvironmentVariables: map[string]string{
						"env_1": "foo",
						"env_2": "182837403930483038",
						"env_3": "true",
						"env_4": "1.00001",
					},
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
					Instances: types.NullInt{
						Value: 10,
						IsSet: true,
					},
					DiskQuota: types.NullByteSizeInMb{
						Value: 1024,
						IsSet: true,
					},
					Memory: types.NullByteSizeInMb{
						Value: 200,
						IsSet: true,
					},
					Routes:             []string{"foo.bar.com", "baz.qux.com"},
					Services:           []string{"service_1", "service_2"},
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				}
			})

			It("creates and writes the manifest to the specified filepath", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  routes:
  - route: foo.bar.com
  - route: baz.qux.com
  services:
  - service_1
  - service_2
  stack: some-stack
  timeout: 120
`))
			})
		})

		Context("when some properties are not provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
				}
			})

			It("does not save them in manifest", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		Context("when the file is a relative path", func() {
			var pwd string

			BeforeEach(func() {
				var err error
				pwd, err = os.Getwd()
				Expect(err).ToNot(HaveOccurred())

				filePath = "./manifest.yml"
				Expect(os.Chdir(tmpDir)).To(Succeed())

				application = Application{
					Name: "app-1",
				}
			})

			AfterEach(func() {
				Expect(os.Chdir(pwd)).To(Succeed())
			})

			It("writes the file in an expanded path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				manifestBytes, err := ioutil.ReadFile(filepath.Join(tmpDir, "manifest.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		Context("when the file already exists", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filePath, []byte(`{}`), 0644)
				Expect(err).ToNot(HaveOccurred())
				application = Application{
					Name: "app-1",
				}
			})

			Context("writes the file", func() {
				It("truncates and writes the manifest to specified filepath", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					manifestBytes, err := ioutil.ReadFile(filePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
				})
			})
		})
	})
})
