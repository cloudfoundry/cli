package command_registry_test

import (
	"strings"

	. "github.com/cloudfoundry/cli/cf/command_registry/fake_command"

	. "github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/cloudfoundry/cli/cf/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRegistry", func() {

	BeforeEach(func() {
		Register(FakeCommand1{})
	})

	Context("i18n", func() {
		It("initialize i18n T() func", func() {
			Ω(T).ToNot(BeNil())
		})
	})

	Context("Register()", func() {
		It("registers a command and it's alias into the Command Registry map", func() {
			Ω(Commands.CommandExists("fake-command2")).To(BeFalse())
			Ω(Commands.CommandExists("fc2")).To(BeFalse())

			Register(FakeCommand2{})

			Ω(Commands.CommandExists("fake-command2")).To(BeTrue())
			Ω(Commands.CommandExists("fc2")).To(BeTrue())
		})
	})

	Describe("Commands", func() {
		Context("CommandExists()", func() {
			It("returns true the command exists in the list", func() {
				Ω(Commands.CommandExists("fake-command")).To(BeTrue())
			})

			It("returns false if the command doesn't exists in the list", func() {
				Ω(Commands.CommandExists("non-exist-cmd")).To(BeFalse())
			})

			It("returns true if the alias exists", func() {
				Ω(Commands.CommandExists("fc1")).To(BeTrue())
			})

			It("returns false if the command name is an empty string", func() {
				Ω(Commands.CommandExists("")).To(BeFalse())
			})
		})

		Context("FindCommand()", func() {
			It("returns the command interface when found", func() {
				cmd := Commands.FindCommand("fake-command")
				Ω(cmd.MetaData().Usage).To(ContainSubstring("Usage of fake-command"))
				Ω(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})

			It("returns the command interface if an command alias is provided", func() {
				cmd := Commands.FindCommand("fc1")
				Ω(cmd.MetaData().Usage).To(ContainSubstring("Usage of fake-command"))
				Ω(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})
		})

		Context("SetCommand()", func() {
			It("replaces the command in registry with command provided", func() {
				updatedCmd := FakeCommand1{Data: "This is new data"}
				oldCmd := Commands.FindCommand("fake-command")
				Ω(oldCmd).ToNot(Equal(updatedCmd))

				Commands.SetCommand(updatedCmd)
				oldCmd = Commands.FindCommand("fake-command")
				Ω(oldCmd).To(Equal(updatedCmd))
			})
		})

		Context("RemoveCommand()", func() {
			It("removes the command in registry with command name provided", func() {
				Commands = NewRegistry()

				Register(FakeCommand1{})
				Register(FakeCommand2{})
				Register(FakeCommand3{})

				Ω(Commands.TotalCommands()).To(Equal(3))
			})
		})

		Context("Metadatas()", func() {
			It("returns all the command's metadata", func() {
				Commands = NewRegistry()

				Register(FakeCommand1{})
				Register(FakeCommand2{})
				Register(FakeCommand3{})

				Ω(len(Commands.Metadatas())).To(Equal(3))
			})
		})

		Context("TotalCommands()", func() {
			It("returns the total number of commands in the registry", func() {
				Ω(Commands.CommandExists("fake-command")).To(BeTrue())

				Commands.RemoveCommand("fake-command")
				Ω(Commands.CommandExists("fake-command")).To(BeFalse())
			})
		})

		Context("MaxCommandNameLength()", func() {
			It("returns the length of the longest command name", func() {
				maxLen := Commands.MaxCommandNameLength()
				Ω(maxLen).To(Equal(len("this-is-a-really-long-command-name-123123123123123123123")))
			})
		})

		Context("CommandUsage()", func() {
			It("prints the name, description and usage of a command", func() {
				o := Commands.CommandUsage("fake-command")
				outputs := strings.Split(o, "\n")
				Ω(outputs).To(BeInDisplayOrder(
					[]string{"NAME:"},
					[]string{"   fake-command", "Description"},
					[]string{"USAGE:"},
				))
			})

			It("prints the flag options", func() {
				o := Commands.CommandUsage("fake-command")
				outputs := strings.Split(o, "\n")
				Ω(outputs).To(BeInDisplayOrder(
					[]string{"NAME:"},
					[]string{"USAGE:"},
					[]string{"OPTIONS:"},
					[]string{"intFlag", "Usage for"},
					[]string{"boolFlag", "Usage for"},
				))
			})

			It("replaces 'CF_NAME' with executable name from os.Arg[0]", func() {
				o := Commands.CommandUsage("fake-command")
				outputs := strings.Split(o, "\n")
				Ω(outputs).To(BeInDisplayOrder(
					[]string{"USAGE:"},
					[]string{"command_registry.test", "Usage of"},
				))
				Consistently(outputs).ShouldNot(ContainSubstrings([]string{"CF_NAME"}))
			})

		})
	})

})
