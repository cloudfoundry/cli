package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildpackRequirement", func() {
	var (
		ui *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	It("succeeds when a buildpack with the given name exists", func() {
		buildpack := models.Buildpack{Name: "my-buildpack"}
		buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameBuildpack: buildpack}

		buildpackReq := NewBuildpackRequirement("my-buildpack", ui, buildpackRepo)

		Expect(buildpackReq.Execute()).To(BeTrue())
		Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
		Expect(buildpackReq.GetBuildpack()).To(Equal(buildpack))
	})

	It("fails when the buildpack cannot be found", func() {
		buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameNotFound: true}

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewBuildpackRequirement("foo", ui, buildpackRepo).Execute()
		})
	})
})
