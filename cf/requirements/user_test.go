package requirements_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakefeatureflagsapi "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserRequirement", func() {
	var (
		ui         *testterm.FakeUI
		userRepo   *fakeapi.FakeUserRepository
		flagRepo   *fakefeatureflagsapi.FakeFeatureFlagRepository
		configRepo core_config.Repository

		userRequirement requirements.UserRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		userRepo = &fakeapi.FakeUserRepository{}
		flagRepo = &fakefeatureflagsapi.FakeFeatureFlagRepository{}
		configRepo = testconfig.NewRepositoryWithDefaults()

		userRequirement = requirements.NewUserRequirement("the-username", ui, userRepo, flagRepo, configRepo)
	})

	Context("when the set_roles_by_username flag is enabled", func() {
		BeforeEach(func() {
			flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: true}, nil)
		})

		Context("when the config version is >=2.37.0", func() {
			BeforeEach(func() {
				configRepo.SetApiVersion("2.37.0")
			})

			It("returns true", func() {
				ok := userRequirement.Execute()
				Expect(ok).To(BeTrue())
			})

			It("requests the set_roles_by_username flag", func() {
				userRequirement.Execute()
				Expect(flagRepo.FindByNameCallCount()).To(Equal(1))
				Expect(flagRepo.FindByNameArgsForCall(0)).To(Equal("set_roles_by_username"))
			})

			It("stores a user with just Username set", func() {
				userRequirement.Execute()
				expectedUser := models.UserFields{Username: "the-username"}
				Expect(userRequirement.GetUser()).To(Equal(expectedUser))
			})
		})

		Context("when the config version is <2.37.0", func() {
			BeforeEach(func() {
				configRepo.SetApiVersion("2.36.0")
			})

			It("tries to find the user in CC", func() {
				userRequirement.Execute()
				Expect(userRepo.FindByUsernameCallCount()).To(Equal(1))
				Expect(userRepo.FindByUsernameArgsForCall(0)).To(Equal("the-username"))
			})

			Context("when a user with the given username can be found", func() {
				var user models.UserFields
				BeforeEach(func() {
					user = models.UserFields{Username: "the-username", Guid: "the-guid"}
					userRepo.FindByUsernameReturns(user, nil)
				})

				It("stores the user that was found", func() {
					userRequirement.Execute()
					Expect(userRequirement.GetUser()).To(Equal(user))
				})
			})

			Context("when call to find the user fails", func() {
				BeforeEach(func() {
					userRepo.FindByUsernameReturns(models.UserFields{}, errors.New("some error"))
				})

				It("panics and prints a failure message", func() {
					Expect(func() { userRequirement.Execute() }).To(Panic())
					Expect(ui.Outputs).To(ConsistOf([]string{"FAILED", "some error"}))
				})
			})
		})
	})

	Context("when the set_roles_by_username flag is disabled", func() {
		BeforeEach(func() {
			flagRepo.FindByNameReturns(models.FeatureFlag{Enabled: true}, nil)
		})

		It("tries to find the user in CC", func() {
			userRequirement.Execute()
			Expect(userRepo.FindByUsernameCallCount()).To(Equal(1))
			Expect(userRepo.FindByUsernameArgsForCall(0)).To(Equal("the-username"))
		})

		Context("when a user with the given username can be found", func() {
			var user models.UserFields
			BeforeEach(func() {
				user = models.UserFields{Username: "the-username", Guid: "the-guid"}
				userRepo.FindByUsernameReturns(user, nil)
			})

			It("stores the user that was found", func() {
				userRequirement.Execute()
				Expect(userRequirement.GetUser()).To(Equal(user))
			})
		})

		Context("when call to find the user fails", func() {
			BeforeEach(func() {
				userRepo.FindByUsernameReturns(models.UserFields{}, errors.New("some error"))
			})

			It("panics and prints a failure message", func() {
				Expect(func() { userRequirement.Execute() }).To(Panic())
				Expect(ui.Outputs).To(ConsistOf([]string{"FAILED", "some error"}))
			})
		})
	})

	Context("when the set_roles_by_username flag cannot be retrieved", func() {
		BeforeEach(func() {
			flagRepo.FindByNameReturns(models.FeatureFlag{}, errors.New("some error"))
		})

		It("panics and prints a failure message", func() {
			Expect(func() { userRequirement.Execute() }).To(Panic())
			Expect(ui.Outputs).To(ConsistOf([]string{"FAILED", "some error"}))
		})
	})
})
