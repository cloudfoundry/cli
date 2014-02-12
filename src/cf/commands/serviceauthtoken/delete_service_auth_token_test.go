package serviceauthtoken_test

import (
	. "cf/commands/serviceauthtoken"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callDeleteServiceAuthToken(args []string, inputs []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("delete-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteServiceAuthTokenFailsWithUsage", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}

		ui := callDeleteServiceAuthToken([]string{}, []string{"Y"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callDeleteServiceAuthToken([]string{"arg1"}, []string{"Y"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callDeleteServiceAuthToken([]string{"arg1", "arg2"}, []string{"Y"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestDeleteServiceAuthTokenRequirements", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}
		args := []string{"arg1", "arg2"}

		reqFactory.LoginSuccess = true
		callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory.LoginSuccess = false
		callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestDeleteServiceAuthToken", func() {

		expectedToken := models.ServiceAuthTokenFields{}
		expectedToken.Label = "a label"
		expectedToken.Provider = "a provider"

		authTokenRepo := &testapi.FakeAuthTokenRepo{
			FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider"}

		ui := callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service auth token as", "my-user"},
			{"OK"},
		})

		Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
		Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
		Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(expectedToken))
	})
	It("TestDeleteServiceAuthTokenWithN", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider"}

		ui := callDeleteServiceAuthToken(args, []string{"N"}, reqFactory, authTokenRepo)

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Are you sure you want to delete", "a label", "a provider"},
		})
		Expect(len(ui.Outputs)).To(Equal(0))
		Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(models.ServiceAuthTokenFields{}))
	})
	It("TestDeleteServiceAuthTokenWithY", func() {

		expectedToken := models.ServiceAuthTokenFields{}
		expectedToken.Label = "a label"
		expectedToken.Provider = "a provider"

		authTokenRepo := &testapi.FakeAuthTokenRepo{
			FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider"}

		ui := callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Are you sure you want to delete", "a label", "a provider"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting"},
			{"OK"},
		})

		Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(expectedToken))
	})
	It("TestDeleteServiceAuthTokenWithForce", func() {

		expectedToken := models.ServiceAuthTokenFields{}
		expectedToken.Label = "a label"
		expectedToken.Provider = "a provider"

		authTokenRepo := &testapi.FakeAuthTokenRepo{
			FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"-f", "a label", "a provider"}
		ui := callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting"},
			{"OK"},
		})

		Expect(authTokenRepo.DeletedServiceAuthTokenFields).To(Equal(expectedToken))
	})
	It("TestDeleteServiceAuthTokenWhenTokenDoesNotExist", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{
			FindByLabelAndProviderApiResponse: net.NewNotFoundApiResponse("not found"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider"}

		ui := callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service auth token as", "my-user"},
			{"OK"},
			{"does not exist"},
		})
	})
	It("TestDeleteServiceAuthTokenFailsWithError", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{
			FindByLabelAndProviderApiResponse: net.NewApiResponseWithMessage("OH NOES"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider"}

		ui := callDeleteServiceAuthToken(args, []string{"Y"}, reqFactory, authTokenRepo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service auth token as", "my-user"},
			{"FAILED"},
			{"OH NOES"},
		})
	})
})
