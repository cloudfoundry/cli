package manifest_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/types"
	. "code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Manifest", func() {
	var manifest string

	Describe("ReadAndInterpolateManifest", func() {
		var (
			pathToManifest   string
			pathsToVarsFiles []string
			vars             []template.VarKV
			apps             []Application
			executeErr       error
		)

		BeforeEach(func() {
			tempFile, err := os.CreateTemp("", "manifest-test-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()
			vars = nil

			pathsToVarsFiles = nil
		})

		AfterEach(func() {
			Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
			for _, path := range pathsToVarsFiles {
				Expect(os.RemoveAll(path)).ToNot(HaveOccurred())
			}
		})

		JustBeforeEach(func() {
			apps, executeErr = ReadAndInterpolateManifest(pathToManifest, pathsToVarsFiles, vars)
		})

		When("the manifest contains NO variables that need interpolation", func() {
			BeforeEach(func() {
				manifest = `---
applications:
- name: app-1
  buildpack: "some-buildpack"
  command: "some-command"
  health-check-http-endpoint: "\\some-endpoint"
  health-check-type: "http"
  instances: 10
  disk_quota: 100M
  docker:
    image: "some-docker-image"
    username: "some-docker-username"
  memory: 200M
  random-route: true
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
  - route: blep.blah.com/boop
  services:
  - service_1
  - service_2
- name: "app-3"
  no-route: true
  env:
    env_1: 'foo'
    env_2: 182837403930483038
    env_3: true
    env_4: 1.00001
- name: "app-4"
  buildpack: null
  command: null
- name: "app-5"
  domain: "some-domain"
  domains:
  - domain_1
  - domain_2
- name: "app-6"
  host: "some-hostname"
  hosts:
  - hostname_1
  - hostname_2
  no-hostname: true
- name: "app-7"
  routes:
  - route: hello.com
  - route: bleep.blah.com
  random-route: true
`

				err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			When("the manifest does not contain deprecated fields", func() {
				It("returns a merged set of applications", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps).To(HaveLen(7))

					Expect(apps[0]).To(Equal(Application{
						Name: "app-1",
						Buildpack: types.FilteredString{
							IsSet: true,
							Value: "some-buildpack",
						},
						Command: types.FilteredString{
							IsSet: true,
							Value: "some-command",
						},
						HealthCheckHTTPEndpoint: `\some-endpoint`,
						HealthCheckType:         "http",
						Instances: types.NullInt{
							Value: 10,
							IsSet: true,
						},
						DiskQuota: types.NullByteSizeInMb{
							Value: 100,
							IsSet: true,
						},
						DockerImage:    "some-docker-image",
						DockerUsername: "some-docker-username",
						Memory: types.NullByteSizeInMb{
							Value: 200,
							IsSet: true,
						},
						RandomRoute:        true,
						StackName:          "some-stack",
						HealthCheckTimeout: 120,
					}))

					Expect(apps[1]).To(Equal(Application{
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
						Routes:   []string{"foo.bar.com", "baz.qux.com", "blep.blah.com/boop"},
						Services: []string{"service_1", "service_2"},
					}))

					Expect(apps[2]).To(Equal(Application{
						Name: "app-3",
						EnvironmentVariables: map[string]string{
							"env_1": "foo",
							"env_2": "182837403930483038",
							"env_3": "true",
							"env_4": "1.00001",
						},
						NoRoute: true,
					}))

					Expect(apps[3]).To(Equal(Application{
						Name: "app-4",
						Buildpack: types.FilteredString{
							IsSet: true,
							Value: "",
						},
						Command: types.FilteredString{
							IsSet: true,
							Value: "",
						},
					}))

					Expect(apps[4].Name).To(Equal("app-5"))
					Expect(apps[4].DeprecatedDomain).ToNot(BeNil())
					Expect(apps[4].DeprecatedDomains).ToNot(BeNil())

					Expect(apps[5].Name).To(Equal("app-6"))
					Expect(apps[5].DeprecatedHost).ToNot(BeNil())
					Expect(apps[5].DeprecatedHosts).ToNot(BeNil())
					Expect(apps[5].DeprecatedNoHostname).ToNot(BeNil())

					Expect(apps[6]).To(Equal(Application{
						Name:        "app-7",
						Routes:      []string{"hello.com", "bleep.blah.com"},
						RandomRoute: true,
					}))
				})
			})

			When("provided deprecated fields", func() {
				When("global fields are provided", func() {
					DescribeTable("raises a GlobalFieldsError",
						func(manifestProperty string, numberOfValues int) {
							tempManifest, err := os.CreateTemp("", "manifest-test-")
							Expect(err).ToNot(HaveOccurred())
							Expect(tempManifest.Close()).ToNot(HaveOccurred())
							manifestPath := tempManifest.Name()
							defer os.RemoveAll(manifestPath)

							if numberOfValues == 1 {
								manifest = fmt.Sprintf("---\n%s: value", manifestProperty)
							} else {
								values := []string{"A", "B"}
								manifest = fmt.Sprintf("---\n%s: [%s]", manifestProperty, strings.Join(values, ","))
							}
							err = os.WriteFile(manifestPath, []byte(manifest), 0666)
							Expect(err).ToNot(HaveOccurred())

							_, err = ReadAndInterpolateManifest(manifestPath, pathsToVarsFiles, vars)
							Expect(err).To(MatchError(GlobalFieldsError{Fields: []string{manifestProperty}}))
						},

						Entry("global buildpack", "buildpack", 1),
						Entry("global command", "command", 1),
						Entry("global disk quota", "disk_quota", 1),
						Entry("global docker", "docker", 1),
						Entry("global domain", "domain", 1),
						Entry("global domains", "domains", 2),
						Entry("global environment variables", "env", 2),
						Entry("global health check HTTP endpoint", "health-check-http-endpoint", 1),
						Entry("global health check timeout", "timeout", 1),
						Entry("global health check type", "health-check-type", 1),
						Entry("global host", "host", 1),
						Entry("global hosts", "hosts", 2),
						Entry("global instances", "instances", 1),
						Entry("global memory", "memory", 1),
						Entry("global name", "name", 1),
						Entry("global no hostname", "no-hostname", 1),
						Entry("global no route", "no-route", 1),
						Entry("global path", "path", 1),
						Entry("global random-route", "random-route", 1),
						Entry("global routes", "routes", 2),
						Entry("global services", "services", 2),
						Entry("global stack", "stack", 1),
					)
				})
			})

			When("inheritance is provided", func() {
				BeforeEach(func() {
					manifest = `---
inherit: "./some-inheritance-file"
applications:
- name: "app-1"
`

					err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("raises an InheritanceFieldError", func() {
					Expect(executeErr).To(MatchError(InheritanceFieldError{}))
				})
			})

			When("the manifest specified a single buildpack", func() {
				BeforeEach(func() {
					manifest = `---
applications:
- name: app-1
  buildpack: "some-buildpack"
  memory: 200M
  instances: 10
`
					err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns a merged set of applications", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps).To(HaveLen(1))

					Expect(apps[0]).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{
							IsSet: true,
							Value: "some-buildpack",
						}),
						"Buildpacks": BeEmpty(),
					}))
				})
			})

			When("the manifest contains buildpacks (plural)", func() {
				BeforeEach(func() {
					manifest = `---
applications:
- name: app-1
  buildpacks:
  - "some-buildpack-1"
  - "some-buildpack-2"
  memory: 200M
  instances: 10
- name: app-2
  buildpacks:
  - "some-other-buildpack-1"
  - "some-other-buildpack-2"
  memory: 2048M
  instances: 0`
					err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns a merged set of applications", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps).To(HaveLen(2))

					Expect(apps[0]).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{
							IsSet: false,
							Value: "",
						}),
						"Buildpacks": ConsistOf("some-buildpack-1", "some-buildpack-2"),
					}))

					Expect(apps[1]).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{
							IsSet: false,
							Value: "",
						}),
						"Buildpacks": ConsistOf(
							"some-other-buildpack-1",
							"some-other-buildpack-2",
						),
					}))
				})
			})

			When("the manifest sets buildpacks to an empty array", func() {
				BeforeEach(func() {
					manifest = `---
applications:
- name: app-1
  buildpacks: []
  memory: 200M
  instances: 10
`
					err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns a merged set of applications", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps).To(HaveLen(1))

					Expect(apps[0]).To(MatchFields(IgnoreExtras, Fields{
						"Buildpack": Equal(types.FilteredString{
							IsSet: false,
							Value: "",
						}),
						"Buildpacks": Equal([]string{}),
					}))
				})
			})

			When("the manifest contains an empty buildpacks attribute", func() {
				BeforeEach(func() {
					manifest = `---
applications:
- name: app-1
  buildpacks:
`
					err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("raises an error", func() {
					Expect(executeErr).ToNot(MatchError(new(EmptyBuildpacksError)))
				})
			})
		})

		When("the manifest contains variables that need interpolation", func() {
			BeforeEach(func() {
				manifest = `---
applications:
- name: ((var1))
  instances: ((var2))
`
				err := os.WriteFile(pathToManifest, []byte(manifest), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			When("only vars files are provided", func() {
				var (
					varsDir string
				)

				BeforeEach(func() {
					var err error
					varsDir, err = os.MkdirTemp("", "vars-test")
					Expect(err).ToNot(HaveOccurred())

					varsFilePath1 := filepath.Join(varsDir, "vars-1")
					err = os.WriteFile(varsFilePath1, []byte("var1: app-1"), 0666)
					Expect(err).ToNot(HaveOccurred())

					varsFilePath2 := filepath.Join(varsDir, "vars-2")
					err = os.WriteFile(varsFilePath2, []byte("var2: 4"), 0666)
					Expect(err).ToNot(HaveOccurred())

					pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
				})

				AfterEach(func() {
					Expect(os.RemoveAll(varsDir)).ToNot(HaveOccurred())
				})

				When("multiple values for the same variable(s) are provided", func() {
					BeforeEach(func() {
						varsFilePath1 := filepath.Join(varsDir, "vars-1")
						err := os.WriteFile(varsFilePath1, []byte("var1: garbageapp\nvar1: app-1\nvar2: 0"), 0666)
						Expect(err).ToNot(HaveOccurred())

						varsFilePath2 := filepath.Join(varsDir, "vars-2")
						err = os.WriteFile(varsFilePath2, []byte("var2: 4"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
					})

					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(apps[0].Name).To(Equal("app-1"))
						Expect(apps[0].Instances).To(Equal(types.NullInt{Value: 4, IsSet: true}))
					})
				})

				When("the provided files exists and contain valid yaml", func() {
					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(apps[0].Name).To(Equal("app-1"))
						Expect(apps[0].Instances).To(Equal(types.NullInt{Value: 4, IsSet: true}))
					})
				})

				When("a variable in the manifest is not provided in the vars file", func() {
					BeforeEach(func() {
						varsFilePath := filepath.Join(varsDir, "vars-1")
						err := os.WriteFile(varsFilePath, []byte("notvar: foo"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = []string{varsFilePath}
					})

					It("returns an error", func() {
						Expect(executeErr.Error()).To(Equal("Expected to find variables: var1, var2"))
					})
				})

				When("the provided file path does not exist", func() {
					BeforeEach(func() {
						pathsToVarsFiles = []string{"garbagepath"}
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(os.IsNotExist(executeErr)).To(BeTrue())
					})
				})

				When("the provided file is not a valid yaml file", func() {
					BeforeEach(func() {
						varsFilePath := filepath.Join(varsDir, "vars-1")
						err := os.WriteFile(varsFilePath, []byte(": bad"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = []string{varsFilePath}
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr).To(MatchError(InvalidYAMLError{
							Err: errors.New("yaml: did not find expected key"),
						}))
					})
				})
			})

			When("only vars are provided", func() {
				BeforeEach(func() {
					vars = []template.VarKV{
						{Name: "var1", Value: "app-1"},
						{Name: "var2", Value: 4},
					}
					manifest = `---
applications:
- name: ((var1))
  instances: ((var2))
`
				})

				It("interpolates the placeholder values", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps[0].Name).To(Equal("app-1"))
					Expect(apps[0].Instances).To(Equal(types.NullInt{Value: 4, IsSet: true}))
				})
			})

			When("vars and vars files are provided", func() {
				var varsFilePath string
				BeforeEach(func() {
					tmp, err := os.CreateTemp("", "util-manifest-varsilfe")
					Expect(err).NotTo(HaveOccurred())
					Expect(tmp.Close()).NotTo(HaveOccurred())

					varsFilePath = tmp.Name()
					err = os.WriteFile(varsFilePath, []byte("var1: app-1\nvar2: 12345"), 0666)
					Expect(err).ToNot(HaveOccurred())

					pathsToVarsFiles = []string{varsFilePath}
					vars = []template.VarKV{
						{Name: "var2", Value: 4},
					}
				})

				AfterEach(func() {
					Expect(os.RemoveAll(varsFilePath)).ToNot(HaveOccurred())
				})

				It("interpolates the placeholder values, prioritizing the vars flag", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(apps[0].Name).To(Equal("app-1"))
					Expect(apps[0].Instances).To(Equal(types.NullInt{Value: 4, IsSet: true}))
				})
			})
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
			tmpDir, err = os.MkdirTemp("", "manifest-test-")
			Expect(err).NotTo(HaveOccurred())
			filePath = filepath.Join(tmpDir, "manifest.yml")
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		JustBeforeEach(func() {
			executeErr = WriteApplicationManifest(application, filePath)
		})

		When("all app properties are provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
					Buildpack: types.FilteredString{
						IsSet: true,
						Value: "some-buildpack",
					},
					Buildpacks: []string{"buildpack1", "buildpack2"},
					Command: types.FilteredString{
						IsSet: true,
						Value: "some-command",
					},
					DockerImage:    "some-docker-image",
					DockerUsername: "some-docker-username",
					DockerPassword: "",
					EnvironmentVariables: map[string]string{
						"env_1": "foo",
						"env_2": "182837403930483038",
						"env_3": "true",
						"env_4": "1.00001",
					},
					HealthCheckHTTPEndpoint: `\some-endpoint`,
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
					NoRoute:            true,
					Routes:             []string{"foo.bar.com", "baz.qux.com", "blep.blah.com/boop"},
					Services:           []string{"service_1", "service_2"},
					StackName:          "some-stack",
					HealthCheckTimeout: 120,
				}
			})

			It("creates and writes the manifest to the specified filepath", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := os.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
  buildpack: some-buildpack
  buildpacks:
  - buildpack1
  - buildpack2
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  no-route: true
  routes:
  - route: foo.bar.com
  - route: baz.qux.com
  - route: blep.blah.com/boop
  services:
  - service_1
  - service_2
  stack: some-stack
  timeout: 120
`))
			})
		})

		When("some properties are not provided", func() {
			BeforeEach(func() {
				application = Application{
					Name: "app-1",
				}
			})

			It("does not save them in manifest", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				manifestBytes, err := os.ReadFile(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		When("the file is a relative path", func() {
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
				manifestBytes, err := os.ReadFile(filepath.Join(tmpDir, "manifest.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
			})
		})

		When("the file already exists", func() {
			BeforeEach(func() {
				err := os.WriteFile(filePath, []byte(`{}`), 0644)
				Expect(err).ToNot(HaveOccurred())
				application = Application{
					Name: "app-1",
				}
			})

			Context("writes the file", func() {
				It("truncates and writes the manifest to specified filepath", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					manifestBytes, err := os.ReadFile(filePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(manifestBytes)).To(Equal(`applications:
- name: app-1
`))
				})
			})
		})
	})
})
