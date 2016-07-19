package resources_test

import (
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application resources", func() {
	var resource *resources.ApplicationResource

	Describe("New Application", func() {
		BeforeEach(func() {
			resource = new(resources.ApplicationResource)
		})

		It("Adds a packageUpdatedAt timestamp", func() {
			err := json.Unmarshal([]byte(`
			{
				"metadata": {
					"guid":"application-1-guid"
				},
				"entity": {
					"package_updated_at": "2013-10-07T16:51:07+00:00"
				}
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			applicationModel := resource.ToModel()
			timestamp, err := time.Parse(eventTimestampFormat, "2013-10-07T16:51:07+00:00")
			Expect(err).ToNot(HaveOccurred())
			Expect(*applicationModel.PackageUpdatedAt).To(Equal(timestamp))
		})
	})

	Describe("NewApplicationEntityFromAppParams", func() {
		var (
			appParams models.AppParams

			diskQuota, memory                 int64
			healthCheckTimeout, instanceCount int
			diego, enableSSH                  bool
			packageUpdatedAt                  time.Time
			appPorts                          []int
			environmentVars                   map[string]interface{}

			buildpackURL,
			command,
			healthCheckType,
			dockerImage,
			name,
			spaceGUID,
			stackGUID,
			state string
		)

		BeforeEach(func() {
			buildpackURL = "buildpack-url"
			command = "command"
			diskQuota = int64(1024)
			environmentVars = map[string]interface{}{
				"foo": "bar",
				"baz": "quux",
			}
			healthCheckType = "none"
			healthCheckTimeout = 5
			dockerImage = "docker-image"
			diego = true
			enableSSH = true
			instanceCount = 5
			memory = int64(2048)
			name = "app-name"
			spaceGUID = "space-guid"
			stackGUID = "stack-guid"
			state = "state"
			packageUpdatedAt = time.Now()
			appPorts = []int{9090, 123}

			appParams = models.AppParams{
				BuildpackURL:       &buildpackURL,
				Command:            &command,
				DiskQuota:          &diskQuota,
				EnvironmentVars:    &environmentVars,
				HealthCheckType:    &healthCheckType,
				HealthCheckTimeout: &healthCheckTimeout,
				DockerImage:        &dockerImage,
				Diego:              &diego,
				EnableSSH:          &enableSSH,
				InstanceCount:      &instanceCount,
				Memory:             &memory,
				Name:               &name,
				SpaceGUID:          &spaceGUID,
				StackGUID:          &stackGUID,
				State:              &state,
				PackageUpdatedAt:   &packageUpdatedAt,
				AppPorts:           &appPorts,
			}
		})

		It("directly assigns some attributes", func() {
			entity := resources.NewApplicationEntityFromAppParams(appParams)
			Expect(*entity.Buildpack).To(Equal(buildpackURL))
			Expect(*entity.Name).To(Equal(name))
			Expect(*entity.SpaceGUID).To(Equal(spaceGUID))
			Expect(*entity.Instances).To(Equal(instanceCount))
			Expect(*entity.Memory).To(Equal(memory))
			Expect(*entity.DiskQuota).To(Equal(diskQuota))
			Expect(*entity.StackGUID).To(Equal(stackGUID))
			Expect(*entity.Command).To(Equal(command))
			Expect(*entity.HealthCheckType).To(Equal(healthCheckType))
			Expect(*entity.HealthCheckTimeout).To(Equal(healthCheckTimeout))
			Expect(*entity.DockerImage).To(Equal(dockerImage))
			Expect(*entity.Diego).To(Equal(diego))
			Expect(*entity.EnableSSH).To(Equal(enableSSH))
			Expect(*entity.PackageUpdatedAt).To(Equal(packageUpdatedAt))
			Expect(*entity.AppPorts).To(Equal(appPorts))
		})

		It("upcases the state", func() {
			entity := resources.NewApplicationEntityFromAppParams(appParams)
			Expect(*entity.State).To(Equal("STATE"))
		})

		It("does not include environment vars when they do not exist in the params", func() {
			appParams.EnvironmentVars = nil
			entity := resources.NewApplicationEntityFromAppParams(appParams)
			Expect(entity.EnvironmentJSON).To(BeNil())
		})
	})
})
