package models_test

import (
	. "code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstance", func() {
	var (
		serviceInstance ServiceInstance
	)

	BeforeEach(func() {
		serviceInstance = ServiceInstance{}
	})

	Describe("isUserProvided", func() {
		When("service instance is of non user provided type", func() {
			BeforeEach(func() {
				serviceInstance.Type = "managed_service_instance"
			})

			It("returns false", func() {
				Expect(serviceInstance.IsUserProvided()).To(BeFalse())
			})
		})

		When("service instance is of user provided type", func() {
			BeforeEach(func() {
				serviceInstance.Type = "user_provided_service_instance"
			})

			It("returns true", func() {
				Expect(serviceInstance.IsUserProvided()).To(BeTrue())
			})
		})

	})
})
