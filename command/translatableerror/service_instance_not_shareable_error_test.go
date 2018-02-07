package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceNotShareableError", func() {
	var serviceInstanceNotShareableErr ServiceInstanceNotShareableError

	BeforeEach(func() {
		serviceInstanceNotShareableErr = ServiceInstanceNotShareableError{}
	})

	Describe("Error", func() {
		Context("when feature flag is disabled, and service broker sharing is disabled", func() {
			BeforeEach(func() {
				serviceInstanceNotShareableErr.FeatureFlagEnabled = false
				serviceInstanceNotShareableErr.ServiceBrokerSharingEnabled = false
			})

			It("returns the appropriate string", func() {
				Expect(serviceInstanceNotShareableErr.Error()).To(Equal(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`))
			})
		})

		Context("when feature flag is enabled, and service broker sharing is disabled", func() {
			BeforeEach(func() {
				serviceInstanceNotShareableErr.FeatureFlagEnabled = true
				serviceInstanceNotShareableErr.ServiceBrokerSharingEnabled = false
			})

			It("returns the appropriate string", func() {
				Expect(serviceInstanceNotShareableErr.Error()).To(Equal("Service instance sharing is disabled for this service."))
			})
		})

		Context("when feature flag is disabled, and service broker sharing is enabled", func() {
			BeforeEach(func() {
				serviceInstanceNotShareableErr.FeatureFlagEnabled = false
				serviceInstanceNotShareableErr.ServiceBrokerSharingEnabled = true
			})

			It("returns the appropriate string", func() {
				Expect(serviceInstanceNotShareableErr.Error()).To(Equal(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`))
			})
		})

		Context("when feature flag is enabled, and service broker sharing is enabled", func() {
			BeforeEach(func() {
				serviceInstanceNotShareableErr.FeatureFlagEnabled = true
				serviceInstanceNotShareableErr.ServiceBrokerSharingEnabled = true
			})

			It("returns unexpected scenario because this error should not be used in this case", func() {
				Expect(serviceInstanceNotShareableErr.Error()).To(Equal("Unexpected ServiceInstanceNotShareableError: service instance is shareable."))
			})
		})
	})
})
