package common_test

import (
	. "code.cloudfoundry.org/cli/command/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("commandList", func() {
	Describe("HasCommand", func() {
		Context("when the command name exists", func() {
			It("returns true", func() {
				Expect(Commands.HasCommand("version")).To(BeTrue())
			})
		})

		Context("when the command name does not exist", func() {
			It("returns false", func() {
				Expect(Commands.HasCommand("does-not-exist")).To(BeFalse())
			})
		})

		Context("when the command name is empty", func() {
			It("returns false", func() {
				Expect(Commands.HasCommand("")).To(BeFalse())
			})
		})
	})

	Describe("HasAlias", func() {
		Context("when the command alias exists", func() {
			It("returns true", func() {
				Expect(Commands.HasAlias("cups")).To(BeTrue())
			})
		})

		Context("when the command alias does not exist", func() {
			It("returns false", func() {
				Expect(Commands.HasAlias("does-not-exist")).To(BeFalse())
			})
		})

		Context("when the command alias is empty", func() {
			It("returns false", func() {
				Expect(Commands.HasAlias("")).To(BeFalse())
			})
		})
	})
})
