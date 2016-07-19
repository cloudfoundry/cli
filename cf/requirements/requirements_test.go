package requirements_test

import (
	. "code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requirements", func() {
	Context("When there are multiple requirements", func() {
		It("executes all the requirements", func() {
			r1 := new(requirementsfakes.FakeRequirement)
			r1.ExecuteReturns(nil)
			r2 := new(requirementsfakes.FakeRequirement)
			r2.ExecuteReturns(nil)

			// SETUP
			requirements := Requirements{
				r1,
				r2,
			}

			// EXECUTE
			err := requirements.Execute()

			// ASSERT
			Expect(err).NotTo(HaveOccurred())
			Expect(r1.ExecuteCallCount()).To(Equal(1))
			Expect(r2.ExecuteCallCount()).To(Equal(1))
		})

		It("returns the first error that occurs", func() {
			disaster := errors.New("OH NO")
			otherDisaster := errors.New("WHAT!")

			r1 := new(requirementsfakes.FakeRequirement)
			r1.ExecuteReturns(disaster)
			r2 := new(requirementsfakes.FakeRequirement)
			r2.ExecuteReturns(otherDisaster)

			// SETUP
			requirements := Requirements{
				r1,
				r2,
			}

			// EXECUTE
			err := requirements.Execute()

			// ASSERT
			Expect(err).To(Equal(disaster))
			Expect(err).NotTo(Equal(otherDisaster))
			Expect(r1.ExecuteCallCount()).To(Equal(1))
			Expect(r2.ExecuteCallCount()).To(Equal(0))
		})
	})
})
