package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildpackRequirement", func() {
	It("succeeds when a buildpack with the given name exists", func() {
		buildpack := models.Buildpack{Name: "my-buildpack"}
		buildpackRepo := &apifakes.OldFakeBuildpackRepository{FindByNameBuildpack: buildpack}

		buildpackReq := NewBuildpackRequirement("my-buildpack", "", buildpackRepo)

		Expect(buildpackReq.Execute()).NotTo(HaveOccurred())
		Expect(buildpackRepo.FindByNameName).To(Equal("my-buildpack"))
		Expect(buildpackReq.GetBuildpack()).To(Equal(buildpack))
	})

	It("fails when the buildpack cannot be found", func() {
		buildpackRepo := &apifakes.OldFakeBuildpackRepository{FindByNameNotFound: true}

		err := NewBuildpackRequirement("foo", "", buildpackRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Buildpack foo not found"))
	})

	It("fails when more than one buildpack is found with the same name and no stack is specified", func() {
		buildpackRepo := &apifakes.OldFakeBuildpackRepository{FindByNameAmbiguous: true}

		err := NewBuildpackRequirement("foo", "", buildpackRepo).Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Multiple buildpacks named foo found"))
	})

	It("finds buildpacks by stack if specified, in addition to name", func() {
		buildpack := models.Buildpack{Name: "my-buildpack", Stack: "my-stack"}
		buildpackRepo := &apifakes.OldFakeBuildpackRepository{FindByNameAndStackBuildpack: buildpack}

		buildpackReq := NewBuildpackRequirement("my-buildpack", "my-stack", buildpackRepo)

		Expect(buildpackReq.Execute()).NotTo(HaveOccurred())
		Expect(buildpackRepo.FindByNameAndStackName).To(Equal("my-buildpack"))
		Expect(buildpackRepo.FindByNameAndStackStack).To(Equal("my-stack"))
		Expect(buildpackReq.GetBuildpack()).To(Equal(buildpack))
	})
})
