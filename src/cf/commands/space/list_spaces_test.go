package space_test

import (
	"cf/api"
	. "cf/commands/space"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callSpaces(args []string, reqFactory *testreq.FakeReqFactory, config configuration.Reader, spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("spaces", args)

	cmd := NewListSpaces(ui, config, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {

	It("TestSpacesRequirements", func() {
		spaceRepo := &testapi.FakeSpaceRepository{}
		config := testconfig.NewRepository()

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callSpaces([]string{}, reqFactory, config, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callSpaces([]string{}, reqFactory, config, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callSpaces([]string{}, reqFactory, config, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestListingSpaces", func() {
		space := models.Space{}
		space.Name = "space1"
		space2 := models.Space{}
		space2.Name = "space2"
		space3 := models.Space{}
		space3.Name = "space3"
		spaceRepo := &testapi.FakeSpaceRepository{
			Spaces: []models.Space{space, space2, space3},
		}

		config := testconfig.NewRepositoryWithDefaults()
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callSpaces([]string{}, reqFactory, config, spaceRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting spaces in org", "my-org", "my-user"},
			{"space1"},
			{"space2"},
			{"space3"},
		})
	})

	It("TestListingSpacesWhenNoSpaces", func() {
		spaceRepo := &testapi.FakeSpaceRepository{
			Spaces: []models.Space{},
		}

		configRepo := testconfig.NewRepositoryWithDefaults()
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callSpaces([]string{}, reqFactory, configRepo, spaceRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting spaces in org", "my-org", "my-user"},
			{"No spaces found"},
		})
	})

})
