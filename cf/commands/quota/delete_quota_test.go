package quota_test

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/quota"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		quotaRepo           *testapi.FakeQuotaRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		quotaRepo = &testapi.FakeQuotaRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		cmd := NewDeleteQuota(ui, configuration.NewRepositoryWithDefaults(), quotaRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("when the user is not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("fails requirements", func() {
			runCommand("my-quota")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails requirements when called without a quota name", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("When the quota provided exists", func() {
			BeforeEach(func() {
				quota := models.QuotaFields{}
				quota.Name = "my-quota"
				quota.Guid = "my-quota-guid"

				quotaRepo.FindByNameReturns.Quota = quota
			})

			It("deletes a quota with a given name when the user confirms", func() {
				ui.Inputs = []string{"y"}

				runCommand("my-quota")
				Expect(quotaRepo.DeleteCalledWith.Guid).To(Equal("my-quota-guid"))

				Expect(ui.Prompts).To(ContainSubstrings(
					[]string{"Really delete the quota", "my-quota"},
				))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting quota", "my-quota", "my-user"},
					[]string{"OK"},
				))
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "my-quota")

				Expect(quotaRepo.DeleteCalledWith.Guid).To(Equal("my-quota-guid"))

				Expect(ui.Prompts).To(BeEmpty())
			})

			It("shows an error when deletion fails", func() {
				quotaRepo.DeleteReturns.Error = errors.New("some error")

				runCommand("-f", "my-quota")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "my-quota"},
					[]string{"FAILED"},
				))
			})
		})

		Context("when finding the quota fails", func() {
			Context("when the quota provided does not exist", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns.Error = errors.NewModelNotFoundError("Quota", "non-existent-quota")
				})

				It("warns the user when that the quota does not exist", func() {
					runCommand("-f", "non-existent-quota")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Deleting", "non-existent-quota"},
						[]string{"OK"},
					))

					Expect(ui.WarnOutputs).To(ContainSubstrings(
						[]string{"non-existent-quota", "does not exist"},
					))
				})
			})

			Context("when other types of error occur", func() {
				BeforeEach(func() {
					quotaRepo.FindByNameReturns.Error = errors.New("some error")
				})

				It("shows an error", func() {
					runCommand("-f", "my-quota")

					Expect(ui.WarnOutputs).ToNot(ContainSubstrings(
						[]string{"my-quota", "does not exist"},
					))

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
					))

				})
			})
		})
	})
})
