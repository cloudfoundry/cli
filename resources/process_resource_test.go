package resources_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Process", func() {
	Describe("MarshalJSON", func() {
		var (
			process      resources.Process
			processBytes []byte
			err          error
		)

		BeforeEach(func() {
			process = resources.Process{}
		})

		JustBeforeEach(func() {
			processBytes, err = process.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		When("instances is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					Instances: types.NullInt{Value: 0, IsSet: true},
				}
			})

			It("sets the instances to be set", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"instances": 0}`))
			})
		})

		When("memory is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					MemoryInMB: types.NullUint64{Value: 0, IsSet: true},
				}
			})

			It("sets the memory to be set", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"memory_in_mb": 0}`))
			})
		})

		When("disk is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					DiskInMB: types.NullUint64{Value: 0, IsSet: true},
				}
			})

			It("sets the disk to be set", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"disk_in_mb": 0}`))
			})
		})

		When("log rate limit is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					LogRateLimitInBPS: types.NullInt{Value: 1024, IsSet: true},
				}
			})

			It("sets the log rate limit to be set", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"log_rate_limit_in_bytes_per_second": 1024}`))
			})
		})
		When("health check type http is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					HealthCheckType:     constant.HTTP,
					HealthCheckEndpoint: "some-endpoint",
				}
			})

			It("sets the health check type to http and has an endpoint", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"http", "data": {"endpoint": "some-endpoint"}}}`))
			})
		})

		When("health check type port is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					HealthCheckType: constant.Port,
				}
			})

			It("sets the health check type to port", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"port", "data": {}}}`))
			})
		})

		When("health check type process is provided", func() {
			BeforeEach(func() {
				process = resources.Process{
					HealthCheckType: constant.Process,
				}
			})

			It("sets the health check type to process", func() {
				Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"process", "data": {}}}`))
			})
		})

		When("process has no fields provided", func() {
			BeforeEach(func() {
				process = resources.Process{}
			})

			It("sets the health check type to process", func() {
				Expect(string(processBytes)).To(MatchJSON(`{}`))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		var (
			process      resources.Process
			processBytes []byte
			err          error
		)
		BeforeEach(func() {
			processBytes = []byte("{}")
		})

		JustBeforeEach(func() {
			err = json.Unmarshal(processBytes, &process)
			Expect(err).ToNot(HaveOccurred())
		})
		When("log rate limit is provided", func() {
			BeforeEach(func() {
				processBytes = []byte(`{"log_rate_limit_in_bytes_per_second": 512}`)
			})

			It("sets the log rate limit", func() {
				Expect(process).To(MatchFields(IgnoreExtras, Fields{
					"LogRateLimitInBPS": Equal(types.NullInt{Value: 512, IsSet: true}),
				}))
			})
		})
		When("health check type http is provided", func() {
			BeforeEach(func() {
				processBytes = []byte(`{"health_check":{"type":"http", "data": {"endpoint": "some-endpoint"}}}`)
			})

			It("sets the health check type to http and has an endpoint", func() {
				Expect(process).To(MatchFields(IgnoreExtras, Fields{
					"HealthCheckType":     Equal(constant.HTTP),
					"HealthCheckEndpoint": Equal("some-endpoint"),
				}))
			})
		})

		When("health check type port is provided", func() {
			BeforeEach(func() {
				processBytes = []byte(`{"health_check":{"type":"port", "data": {"endpoint": null}}}`)
			})

			It("sets the health check type to port", func() {
				Expect(process).To(MatchFields(IgnoreExtras, Fields{
					"HealthCheckType": Equal(constant.Port),
				}))
			})
		})

		When("health check type process is provided", func() {
			BeforeEach(func() {
				processBytes = []byte(`{"health_check":{"type":"process", "data": {"endpoint": null}}}`)
			})

			It("sets the health check type to process", func() {
				Expect(process).To(MatchFields(IgnoreExtras, Fields{
					"HealthCheckType": Equal(constant.Process),
				}))
			})
		})
	})
})
