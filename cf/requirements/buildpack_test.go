package requirements_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildpackRequirement", func() {
	It("succeeds when a buildpack with the given name exists", func() {
		buildpack := models.Buildpack{Name: "my-buildpack"}
		buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameBuildpack: buildpack}

		buildpackReq := NewBuildpackRequirement("my-buildpack", buildpackRepo)

		Expect(buildpackReq.Execute()).NotTo(HaveOccurred())
		Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
		Expect(buildpackReq.GetBuildpack()).To(Equal(buildpack))
	})

	It("fails when the buildpack cannot be found", func() {
		buildpackRepo := &testapi.FakeBuildpackRepository{FindByNameNotFound: true}

		err := NewBuildpackRequirement("foo", buildpackRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Buildpack foo not found"))
	})
})
