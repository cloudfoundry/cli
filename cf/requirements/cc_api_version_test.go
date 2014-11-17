package requirements_test

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CcApiVersion", func() {
	var (
		ui     *testterm.FakeUI
		config core_config.Repository
		req    CCApiVersionRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
	})

	Describe("success", func() {
		Describe("when the cc api version has only major version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 0, 0)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc api version has only major and minor versions", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 1, 0)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc api version has three dots", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.1.1.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 1, 1)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc major version is trash", func() {
			BeforeEach(func() {
				config.SetApiVersion("garbage.1.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 0, 1, 1)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc minor version is trash", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.garbage.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 0, 1)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc patch version is trash", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.1.garbage")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 1, 0)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when cc major version is greater than the required major version", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.0.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 0, 0)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when cc minor version is greater than the require minor version", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.1.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 0, 0)

			})
			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when cc patch version is greater than the required patch version", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.1.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 1, 0)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc major, minor, and patch are equal to the required versions", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.1.1")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 1, 1)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the cc major version is not higher than require and the minor and patch are", func() {
			BeforeEach(func() {
				config.SetApiVersion("9.2.2")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 9, 9)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})

		Describe("when the major and minor versions are higher than required and the patch is not", func() {
			BeforeEach(func() {
				config.SetApiVersion("9.9.2")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 2, 9)
			})

			It("should pass", func() {
				Expect(req.Execute()).To(BeTrue())
			})
		})
	})

	Describe("failure", func() {
		Describe("when cc major version is less than the required major version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.0.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 0, 0)
			})

			It("should fail", func() {
				Expect(req.Execute()).To(BeFalse())
			})
		})

		Describe("when cc minor version is less than the required minor version", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.0.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 1, 0)
			})

			It("should fail", func() {
				Expect(req.Execute()).To(BeFalse())
			})
		})

		Describe("when cc patch version is less than the required patch version", func() {
			BeforeEach(func() {
				config.SetApiVersion("2.1.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 2, 1, 1)
			})

			It("should fail", func() {
				Expect(req.Execute()).To(BeFalse())
			})
		})

		Describe("output", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.1.0")
				req = NewCCApiVersionRequirement(ui, config, "command-name", 1, 1, 1)
				req.Execute()
			})

			It("should write to the ui", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{fmt.Sprintf("Current CF CLI version %s", cf.Version)},
					[]string{"Current CF API version 1.1.0"},
					[]string{"To use the command-name feature, you need to upgrade the CF API to at least 1.1.1"},
				))
			})
		})
	})
})
