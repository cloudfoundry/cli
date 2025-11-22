package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	It("stores buildpack information", func() {
		position := 1
		enabled := true
		locked := false

		buildpack := models.Buildpack{
			Guid:     "buildpack-guid",
			Name:     "ruby_buildpack",
			Position: &position,
			Enabled:  &enabled,
			Key:      "buildpack-key",
			Filename: "ruby_buildpack-v1.0.0.zip",
			Locked:   &locked,
		}

		Expect(buildpack.Guid).To(Equal("buildpack-guid"))
		Expect(buildpack.Name).To(Equal("ruby_buildpack"))
		Expect(*buildpack.Position).To(Equal(1))
		Expect(*buildpack.Enabled).To(BeTrue())
		Expect(buildpack.Key).To(Equal("buildpack-key"))
		Expect(buildpack.Filename).To(Equal("ruby_buildpack-v1.0.0.zip"))
		Expect(*buildpack.Locked).To(BeFalse())
	})

	It("handles nil position", func() {
		buildpack := models.Buildpack{
			Guid:     "buildpack-guid",
			Name:     "test_buildpack",
			Position: nil,
		}

		Expect(buildpack.Position).To(BeNil())
	})

	It("handles nil enabled flag", func() {
		buildpack := models.Buildpack{
			Guid:    "buildpack-guid",
			Name:    "test_buildpack",
			Enabled: nil,
		}

		Expect(buildpack.Enabled).To(BeNil())
	})

	It("handles nil locked flag", func() {
		buildpack := models.Buildpack{
			Guid:   "buildpack-guid",
			Name:   "test_buildpack",
			Locked: nil,
		}

		Expect(buildpack.Locked).To(BeNil())
	})

	It("handles disabled buildpack", func() {
		enabled := false
		buildpack := models.Buildpack{
			Guid:    "buildpack-guid",
			Name:    "disabled_buildpack",
			Enabled: &enabled,
		}

		Expect(*buildpack.Enabled).To(BeFalse())
	})

	It("handles locked buildpack", func() {
		locked := true
		buildpack := models.Buildpack{
			Guid:   "buildpack-guid",
			Name:   "locked_buildpack",
			Locked: &locked,
		}

		Expect(*buildpack.Locked).To(BeTrue())
	})

	It("handles different positions", func() {
		pos1 := 1
		pos2 := 10
		pos3 := 99

		buildpack1 := models.Buildpack{Position: &pos1}
		buildpack2 := models.Buildpack{Position: &pos2}
		buildpack3 := models.Buildpack{Position: &pos3}

		Expect(*buildpack1.Position).To(Equal(1))
		Expect(*buildpack2.Position).To(Equal(10))
		Expect(*buildpack3.Position).To(Equal(99))
	})

	It("handles empty filename", func() {
		buildpack := models.Buildpack{
			Guid:     "buildpack-guid",
			Name:     "test_buildpack",
			Filename: "",
		}

		Expect(buildpack.Filename).To(BeEmpty())
	})

	It("handles empty key", func() {
		buildpack := models.Buildpack{
			Guid: "buildpack-guid",
			Name: "test_buildpack",
			Key:  "",
		}

		Expect(buildpack.Key).To(BeEmpty())
	})

	It("stores different buildpack names", func() {
		bp1 := models.Buildpack{Name: "ruby_buildpack"}
		bp2 := models.Buildpack{Name: "nodejs_buildpack"}
		bp3 := models.Buildpack{Name: "python_buildpack"}
		bp4 := models.Buildpack{Name: "custom-buildpack"}

		Expect(bp1.Name).To(Equal("ruby_buildpack"))
		Expect(bp2.Name).To(Equal("nodejs_buildpack"))
		Expect(bp3.Name).To(Equal("python_buildpack"))
		Expect(bp4.Name).To(Equal("custom-buildpack"))
	})
})
