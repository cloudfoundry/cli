package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	Describe("Unmarshal", func() {
		var (
			emptyMap    map[string]interface{}
			rawYAML     []byte
			application Application
			executeErr  error
		)

		BeforeEach(func() {
			emptyMap = make(map[string]interface{})
		})

		JustBeforeEach(func() {
			application = Application{}
			executeErr = yaml.Unmarshal(rawYAML, &application)
		})

		Context("when a name is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
name: spark
`)
			})

			It("unmarshals the name", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Name).To(Equal("spark"))
			})
		})

		When("a disk quota is provided", func() {
			When("it has a hyphen (`disk-quota`)", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
disk-quota: 5G
`)
				})

				It("unmarshals the disk quota", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.DiskQuota).To(Equal("5G"))
				})
			})
			When("it has an underscore (`disk_quota`)", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
disk_quota: 5G
`)
				})

				It("unmarshals the disk quota because we maintain backwards-compatibility with the old style", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.DiskQuota).To(Equal("5G"))
				})
				It("doesn't leave a second version of `disk_quota` in the remaining manifest fields", func() {
					Expect(application.RemainingManifestFields["disk_quota"]).To(BeNil())
				})
			})
			When("it has an underscore but it isn't a string", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
disk_quota: [1]
`)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("`disk_quota` must be a string"))
				})
			})
			When("both underscore and hyphen versions are present", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
disk_quota: 5G
disk-quota: 6G
`)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("cannot define both `disk_quota` and `disk-quota`"))
				})
			})
		})

		Context("when default-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
default-route: true
`)
			})

			It("unmarshals the name", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.DefaultRoute).To(BeTrue())
			})
		})

		Context("when a path is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
path: /my/path
`)
			})

			It("unmarshals the path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Path).To(Equal("/my/path"))
			})
		})

		Context("when a docker map is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
docker:
  image: some-image
  username: some-username
`)
			})

			It("unmarshals the docker properties", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Docker.Image).To(Equal("some-image"))
				Expect(application.Docker.Username).To(Equal("some-username"))
			})
		})

		Context("when no-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
no-route: true
`)
			})

			It("unmarshals the no-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.NoRoute).To(BeTrue())
			})
		})

		Context("when random-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
random-route: true
`)
			})

			It("unmarshals the random-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.RandomRoute).To(BeTrue())
			})
		})

		Context("when buildpacks is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
buildpacks:
- ruby_buildpack
- java_buildpack
`)
			})

			It("unmarshals the buildpacks property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.RemainingManifestFields["buildpacks"]).To(ConsistOf("ruby_buildpack", "java_buildpack"))
			})
		})

		Context("when stack is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
stack: cflinuxfs3
`)
			})

			It("unmarshals the stack property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Stack).To(Equal("cflinuxfs3"))
			})
		})

		Context("when an unknown field is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
unknown-key: 2
`)
			})

			It("unmarshals the unknown field to a map", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				mapVal, ok := application.RemainingManifestFields["unknown-key"]
				Expect(ok).To(BeTrue())

				mapValAsInt, ok := mapVal.(int)
				Expect(ok).To(BeTrue())

				Expect(mapValAsInt).To(Equal(2))
			})
		})

		Context("when Processes are provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
processes: []
`)
			})

			It("unmarshals the processes property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Processes).To(Equal([]Process{}))
			})
		})

		Context("process-level configuration", func() {
			Context("the Type command is always provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- type: web
`)
				})

				It("unmarshals the processes property with the type", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{Type: "web", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when the start command is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- command: /bin/python
`)
				})

				It("unmarshals the processes property with the start command", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes[0].RemainingManifestFields["command"]).To(Equal("/bin/python"))
				})
			})

			Context("when a disk quota is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- disk_quota: 5GB
`)
				})

				It("unmarshals the processes property with the disk quota", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{DiskQuota: "5GB", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when a health check endpoint is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- health-check-http-endpoint: https://localhost
`)
				})

				It("unmarshals the processes property with the health check endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{HealthCheckEndpoint: "https://localhost", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when a health check timeout is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- timeout: 42
`)
				})

				It("unmarshals the processes property with the health check endpoint", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{HealthCheckTimeout: 42, RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when a health check type is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- health-check-type: http
`)
				})

				It("unmarshals the processes property with the health check type", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{HealthCheckType: "http", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when a memory limit is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- memory: 512M
`)
				})

				It("unmarshals the processes property with the memory limit", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{Memory: "512M", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when instances are provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- instances: 4
`)
				})

				It("unmarshals the processes property with instances", func() {
					a := 4
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{Instances: &a, RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when a log rate limit is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- log-rate-limit-per-second: 512M
`)
				})

				It("unmarshals the processes property with the log rate limit", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(application.Processes).To(Equal([]Process{
						{LogRateLimit: "512M", RemainingManifestFields: emptyMap},
					}))
				})
			})

			Context("when an unknown field is provided", func() {
				BeforeEach(func() {
					rawYAML = []byte(`---
processes:
- unknown-key: 2
`)
				})

				It("unmarshals the unknown field to a map", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					mapVal, ok := application.Processes[0].RemainingManifestFields["unknown-key"]
					Expect(ok).To(BeTrue())

					mapValAsInt, ok := mapVal.(int)
					Expect(ok).To(BeTrue())

					Expect(mapValAsInt).To(Equal(2))
				})
			})
		})

		Context("marshalling & unmarshalling fields with special-cased empty values", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
name: ""
command: null
buildpacks: []
processes:
- type: web
  command: null
`)
			})

			It("preserves the values as-written", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				remarshalledYaml, err := yaml.Marshal(&application)
				Expect(err).NotTo(HaveOccurred())

				Expect(remarshalledYaml).To(MatchYAML(rawYAML))
			})
		})

		Context("when a log rate limit is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
log-rate-limit-per-second: 5K
`)
			})

			It("unmarshals the log rate limit", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.LogRateLimit).To(Equal("5K"))
			})
		})
	})

	Describe("SetStartCommand", func() {
		var (
			app     Application
			command string
		)

		BeforeEach(func() {
			app = Application{}
			command = "./start.sh"
		})

		JustBeforeEach(func() {
			app.SetStartCommand(command)
		})

		When("the remaining fields map exists", func() {
			BeforeEach(func() {
				app.RemainingManifestFields = map[string]interface{}{}
			})

			It("sets the start command in the map", func() {
				Expect(app.RemainingManifestFields["command"]).To(Equal("./start.sh"))
			})

			When("the command is nil", func() {
				BeforeEach(func() {
					command = ""
				})

				It("sets the start command to nil in the map", func() {
					Expect(app.RemainingManifestFields["command"]).To(BeNil())
				})
			})
		})

		When("the remaining fields map does not exist", func() {
			It("sets the start command in the map", func() {
				Expect(app.RemainingManifestFields["command"]).To(Equal("./start.sh"))
			})
		})
	})

	Describe("SetBuildpacks", func() {
		var (
			app        Application
			buildpacks []string
		)

		BeforeEach(func() {
			app = Application{}
			buildpacks = []string{"bp1", "bp2"}
		})

		JustBeforeEach(func() {
			app.SetBuildpacks(buildpacks)
		})

		When("the remaining fields map exists", func() {
			BeforeEach(func() {
				app.RemainingManifestFields = map[string]interface{}{}
			})

			It("sets the buildpacks in the map", func() {
				Expect(app.RemainingManifestFields["buildpacks"]).To(ConsistOf("bp1", "bp2"))
			})

			When("buildpacks is empty", func() {
				BeforeEach(func() {
					buildpacks = []string{}
				})

				It("sets the buildpacks to empty in the map", func() {
					Expect(app.RemainingManifestFields["buildpacks"]).To(BeEmpty())
				})
			})
		})

		When("the remaining fields map does not exist", func() {
			It("sets the buildpacks in the map", func() {
				Expect(app.RemainingManifestFields["buildpacks"]).To(ConsistOf("bp1", "bp2"))
			})
		})
	})

	Describe("HasBuildpacks", func() {
		var (
			app        Application
			buildpacks []string
		)

		When("the app has buildpacks", func() {
			BeforeEach(func() {
				buildpacks = []string{"bp1", "bp2"}
				app = Application{RemainingManifestFields: map[string]interface{}{"buildpacks": buildpacks}}
			})

			It("returns true", func() {
				Expect(app.HasBuildpacks()).To(BeTrue())
			})
		})

		When("the app does not have buildpacks", func() {
			BeforeEach(func() {
				app = Application{}
			})

			It("returns false", func() {
				Expect(app.HasBuildpacks()).To(BeFalse())
			})
		})
	})
})
