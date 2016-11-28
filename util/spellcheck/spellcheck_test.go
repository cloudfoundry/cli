package spellcheck_test

import (
	. "code.cloudfoundry.org/cli/util/spellcheck"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spellcheck", func() {
	var commandSuggester CommandSuggester
	BeforeEach(func() {
		existingCmds := []string{"fake-command", "fake-command2", "help"}
		commandSuggester = NewCommandSuggester(existingCmds)
	})

	Context("when there is no input", func() {
		It("returns empty slice", func() {
			Expect(commandSuggester.Recommend("")).To(Equal([]string{}))
		})
	})

	Context("when there is an input", func() {
		It("returns recommendations", func() {
			Expect(commandSuggester.Recommend("hlp")).To(Equal([]string{"help"}))
		})
	})
})
