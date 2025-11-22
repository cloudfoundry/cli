package models_test

import (
	"time"

	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Additional Models", func() {
	Describe("Stack", func() {
		It("stores stack information", func() {
			stack := models.Stack{
				Guid:        "stack-guid",
				Name:        "cflinuxfs3",
				Description: "Cloud Foundry Linux-based filesystem",
			}

			Expect(stack.Guid).To(Equal("stack-guid"))
			Expect(stack.Name).To(Equal("cflinuxfs3"))
			Expect(stack.Description).To(Equal("Cloud Foundry Linux-based filesystem"))
		})

		It("handles empty description", func() {
			stack := models.Stack{
				Guid: "stack-guid",
				Name: "custom-stack",
			}

			Expect(stack.Description).To(BeEmpty())
		})

		It("stores different stack names", func() {
			stack1 := models.Stack{Name: "cflinuxfs3"}
			stack2 := models.Stack{Name: "cflinuxfs4"}
			stack3 := models.Stack{Name: "windows2016"}

			Expect(stack1.Name).To(Equal("cflinuxfs3"))
			Expect(stack2.Name).To(Equal("cflinuxfs4"))
			Expect(stack3.Name).To(Equal("windows2016"))
		})
	})

	Describe("SecurityGroupFields", func() {
		It("stores security group fields", func() {
			rules := []map[string]interface{}{
				{"protocol": "tcp", "destination": "10.0.0.0/8"},
				{"protocol": "udp", "destination": "192.168.1.0/24"},
			}

			sg := models.SecurityGroupFields{
				Name:     "my-security-group",
				Guid:     "sg-guid",
				SpaceUrl: "/v2/spaces/space-guid",
				Rules:    rules,
			}

			Expect(sg.Name).To(Equal("my-security-group"))
			Expect(sg.Guid).To(Equal("sg-guid"))
			Expect(sg.SpaceUrl).To(Equal("/v2/spaces/space-guid"))
			Expect(len(sg.Rules)).To(Equal(2))
			Expect(sg.Rules[0]["protocol"]).To(Equal("tcp"))
		})

		It("handles empty rules", func() {
			sg := models.SecurityGroupFields{
				Name:  "empty-sg",
				Rules: []map[string]interface{}{},
			}

			Expect(len(sg.Rules)).To(Equal(0))
		})
	})

	Describe("SecurityGroupParams", func() {
		It("stores security group parameters for API", func() {
			rules := []map[string]interface{}{
				{"protocol": "tcp", "ports": "80,443"},
			}

			params := models.SecurityGroupParams{
				Name:  "web-sg",
				Guid:  "sg-guid",
				Rules: rules,
			}

			Expect(params.Name).To(Equal("web-sg"))
			Expect(params.Guid).To(Equal("sg-guid"))
			Expect(len(params.Rules)).To(Equal(1))
		})
	})

	Describe("SecurityGroup", func() {
		It("embeds SecurityGroupFields", func() {
			sg := models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name: "my-sg",
					Guid: "sg-guid",
				},
			}

			Expect(sg.Name).To(Equal("my-sg"))
			Expect(sg.Guid).To(Equal("sg-guid"))
		})

		It("has associated spaces", func() {
			sg := models.SecurityGroup{
				SecurityGroupFields: models.SecurityGroupFields{
					Name: "my-sg",
				},
				Spaces: []models.Space{
					{SpaceFields: models.SpaceFields{Name: "space1"}},
					{SpaceFields: models.SpaceFields{Name: "space2"}},
				},
			}

			Expect(len(sg.Spaces)).To(Equal(2))
			Expect(sg.Spaces[0].Name).To(Equal("space1"))
		})
	})

	Describe("InstanceState", func() {
		It("defines instance state constants", func() {
			Expect(models.InstanceStarting).To(Equal(models.InstanceState("starting")))
			Expect(models.InstanceRunning).To(Equal(models.InstanceState("running")))
			Expect(models.InstanceFlapping).To(Equal(models.InstanceState("flapping")))
			Expect(models.InstanceDown).To(Equal(models.InstanceState("down")))
			Expect(models.InstanceCrashed).To(Equal(models.InstanceState("crashed")))
		})
	})

	Describe("AppInstanceFields", func() {
		It("stores app instance information", func() {
			now := time.Now()
			instance := models.AppInstanceFields{
				State:     models.InstanceRunning,
				Details:   "instance details",
				Since:     now,
				CpuUsage:  45.5,
				DiskQuota: 1073741824, // 1GB in bytes
				DiskUsage: 536870912,  // 512MB
				MemQuota:  536870912,  // 512MB
				MemUsage:  268435456,  // 256MB
			}

			Expect(instance.State).To(Equal(models.InstanceRunning))
			Expect(instance.Details).To(Equal("instance details"))
			Expect(instance.Since).To(Equal(now))
			Expect(instance.CpuUsage).To(Equal(45.5))
			Expect(instance.DiskQuota).To(Equal(int64(1073741824)))
			Expect(instance.DiskUsage).To(Equal(int64(536870912)))
			Expect(instance.MemQuota).To(Equal(int64(536870912)))
			Expect(instance.MemUsage).To(Equal(int64(268435456)))
		})

		It("handles different instance states", func() {
			starting := models.AppInstanceFields{State: models.InstanceStarting}
			running := models.AppInstanceFields{State: models.InstanceRunning}
			crashed := models.AppInstanceFields{State: models.InstanceCrashed}

			Expect(starting.State).To(Equal(models.InstanceStarting))
			Expect(running.State).To(Equal(models.InstanceRunning))
			Expect(crashed.State).To(Equal(models.InstanceCrashed))
		})

		It("handles zero CPU usage", func() {
			instance := models.AppInstanceFields{
				CpuUsage: 0.0,
			}

			Expect(instance.CpuUsage).To(Equal(0.0))
		})

		It("handles high CPU usage", func() {
			instance := models.AppInstanceFields{
				CpuUsage: 99.9,
			}

			Expect(instance.CpuUsage).To(Equal(99.9))
		})
	})

	Describe("SpaceQuota", func() {
		It("stores space quota information", func() {
			quota := models.SpaceQuota{
				Guid:                    "quota-guid",
				Name:                    "space-quota",
				MemoryLimit:             2048,
				InstanceMemoryLimit:     1024,
				RoutesLimit:             10,
				ServicesLimit:           5,
				NonBasicServicesAllowed: true,
				OrgGuid:                 "org-guid",
			}

			Expect(quota.Guid).To(Equal("quota-guid"))
			Expect(quota.Name).To(Equal("space-quota"))
			Expect(quota.MemoryLimit).To(Equal(int64(2048)))
			Expect(quota.InstanceMemoryLimit).To(Equal(int64(1024)))
			Expect(quota.RoutesLimit).To(Equal(10))
			Expect(quota.ServicesLimit).To(Equal(5))
			Expect(quota.NonBasicServicesAllowed).To(BeTrue())
			Expect(quota.OrgGuid).To(Equal("org-guid"))
		})

		It("handles unlimited instance memory", func() {
			quota := models.SpaceQuota{
				Name:                "unlimited-quota",
				InstanceMemoryLimit: -1,
			}

			Expect(quota.InstanceMemoryLimit).To(Equal(int64(-1)))
		})

		It("handles zero limits", func() {
			quota := models.SpaceQuota{
				Name:          "zero-quota",
				RoutesLimit:   0,
				ServicesLimit: 0,
			}

			Expect(quota.RoutesLimit).To(Equal(0))
			Expect(quota.ServicesLimit).To(Equal(0))
		})
	})

	Describe("FeatureFlag", func() {
		It("stores feature flag information", func() {
			flag := models.FeatureFlag{
				Name:         "user_org_creation",
				Enabled:      true,
				ErrorMessage: "",
			}

			Expect(flag.Name).To(Equal("user_org_creation"))
			Expect(flag.Enabled).To(BeTrue())
			Expect(flag.ErrorMessage).To(BeEmpty())
		})

		It("stores disabled flag", func() {
			flag := models.FeatureFlag{
				Name:         "diego_docker",
				Enabled:      false,
				ErrorMessage: "Docker is disabled",
			}

			Expect(flag.Enabled).To(BeFalse())
			Expect(flag.ErrorMessage).To(Equal("Docker is disabled"))
		})

		It("handles different flag names", func() {
			flag1 := models.FeatureFlag{Name: "private_domain_creation"}
			flag2 := models.FeatureFlag{Name: "app_bits_upload"}
			flag3 := models.FeatureFlag{Name: "service_instance_sharing"}

			Expect(flag1.Name).To(Equal("private_domain_creation"))
			Expect(flag2.Name).To(Equal("app_bits_upload"))
			Expect(flag3.Name).To(Equal("service_instance_sharing"))
		})
	})

	Describe("EnvironmentVariable", func() {
		It("stores environment variable", func() {
			env := models.EnvironmentVariable{
				Name:  "DATABASE_URL",
				Value: "postgres://localhost:5432/mydb",
			}

			Expect(env.Name).To(Equal("DATABASE_URL"))
			Expect(env.Value).To(Equal("postgres://localhost:5432/mydb"))
		})

		It("handles empty value", func() {
			env := models.EnvironmentVariable{
				Name:  "EMPTY_VAR",
				Value: "",
			}

			Expect(env.Name).To(Equal("EMPTY_VAR"))
			Expect(env.Value).To(BeEmpty())
		})

		It("handles special characters in value", func() {
			env := models.EnvironmentVariable{
				Name:  "SPECIAL",
				Value: "value=with&special!chars@#$",
			}

			Expect(env.Value).To(Equal("value=with&special!chars@#$"))
		})

		It("stores different variable names", func() {
			env1 := models.EnvironmentVariable{Name: "PORT"}
			env2 := models.EnvironmentVariable{Name: "NODE_ENV"}
			env3 := models.EnvironmentVariable{Name: "API_KEY"}

			Expect(env1.Name).To(Equal("PORT"))
			Expect(env2.Name).To(Equal("NODE_ENV"))
			Expect(env3.Name).To(Equal("API_KEY"))
		})
	})
})
