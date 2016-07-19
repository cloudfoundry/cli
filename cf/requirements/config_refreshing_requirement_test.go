package requirements_test

import (
	. "code.cloudfoundry.org/cli/cf/requirements"

	"errors"

	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigRefreshingRequirement", func() {
	var (
		r                     Requirement
		underlyingRequirement *requirementsfakes.FakeRequirement
		configRefresher       *requirementsfakes.FakeConfigRefresher
	)

	BeforeEach(func() {
		underlyingRequirement = new(requirementsfakes.FakeRequirement)
		configRefresher = new(requirementsfakes.FakeConfigRefresher)
		r = NewConfigRefreshingRequirement(underlyingRequirement, configRefresher)
	})

	Describe("Execute", func() {
		It("tries to execute the underlying requirement", func() {
			underlyingRequirement.ExecuteReturns(nil)

			r.Execute()

			Expect(underlyingRequirement.ExecuteCallCount()).To(Equal(1))
			Expect(configRefresher.RefreshCallCount()).To(Equal(0))
		})

		Context("when the underlying requirement fails", func() {
			It("refreshes the config", func() {
				underlyingRequirement.ExecuteReturns(errors.New("TERRIBLE THINGS"))

				r.Execute()

				Expect(configRefresher.RefreshCallCount()).To(Equal(1))
			})

			It("returns the value of calling execute on the underlying requirement again", func() {
				var count int
				disaster := errors.New("TERRIBLE THINGS")
				secondaryDisaster := errors.New("REALLY TERRIBLE THINGS")

				underlyingRequirement.ExecuteStub = func() error {
					if count == 0 {
						count++
						return disaster
					}
					return secondaryDisaster
				}

				err := r.Execute()

				Expect(underlyingRequirement.ExecuteCallCount()).To(Equal(2))
				Expect(err).To(Equal(secondaryDisaster))
			})

			Context("if config refresh fails", func() {
				It("returns the error", func() {
					underlyingRequirement.ExecuteReturns(errors.New("TERRIBLE THINGS"))
					oops := errors.New("Can't get things")
					configRefresher.RefreshReturns(nil, oops)

					err := r.Execute()
					Expect(err).To(Equal(oops))

					Expect(underlyingRequirement.ExecuteCallCount()).To(Equal(1))
				})
			})
		})
	})
})
