package command_registry_test

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"

	. "github.com/cloudfoundry/cli/cf/command_registry/fake_command"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/cloudfoundry/cli/cf/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRegistry", func() {
	BeforeEach(func() {
		command_registry.Commands = command_registry.NewRegistry() // because other tests load all the commands into the registry
	})

	Context("i18n", func() {
		It("initialize i18n T() func", func() {
			Expect(T).ToNot(BeNil())
		})
	})

	Describe("Register()", func() {
		AfterEach(func() {
			command_registry.Commands.RemoveCommand("fake-command2")
		})

		It("registers a command and its alias into the Command Registry map", func() {
			Expect(command_registry.Commands.CommandExists("fake-command2")).To(BeFalse())
			Expect(command_registry.Commands.CommandExists("fc2")).To(BeFalse())

			command_registry.Register(FakeCommand2{})

			Expect(command_registry.Commands.CommandExists("fake-command2")).To(BeTrue())
			Expect(command_registry.Commands.CommandExists("fc2")).To(BeTrue())
		})
	})

	Describe("CommandExists()", func() {
		Context("when the command has been registered", func() {
			BeforeEach(func() {
				command_registry.Register(FakeCommand1{})
			})

			AfterEach(func() {
				command_registry.Commands.RemoveCommand("fake-command")
			})

			It("returns true the command exists in the list", func() {
				Expect(command_registry.Commands.CommandExists("fake-command")).To(BeTrue())
			})

			It("returns true if the alias exists", func() {
				Expect(command_registry.Commands.CommandExists("fc1")).To(BeTrue())
			})
		})

		It("returns false when the command has not been registered", func() {
			Expect(command_registry.Commands.CommandExists("non-exist-cmd")).To(BeFalse())
		})

		It("returns false if the command name is an empty string", func() {
			Expect(command_registry.Commands.CommandExists("")).To(BeFalse())
		})
	})

	Describe("FindCommand()", func() {
		Context("when the command has been registered", func() {
			BeforeEach(func() {
				command_registry.Register(FakeCommand1{})
			})

			AfterEach(func() {
				command_registry.Commands.RemoveCommand("fake-command")
			})

			It("returns the command when the command's name is given", func() {
				cmd := command_registry.Commands.FindCommand("fake-command")
				Expect(cmd.MetaData().Usage).To(ContainSubstring("Usage of fake-command"))
				Expect(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})

			It("returns the command when the command's alias is given", func() {
				cmd := command_registry.Commands.FindCommand("fc1")
				Expect(cmd.MetaData().Usage).To(ContainSubstring("Usage of fake-command"))
				Expect(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})
		})

		It("returns nil when the command has not been registered", func() {
			cmd := command_registry.Commands.FindCommand("fake-command")
			Expect(cmd).To(BeNil())
		})
	})

	Describe("SetCommand()", func() {
		It("replaces the command in registry with command provided", func() {
			updatedCmd := FakeCommand1{Data: "This is new data"}
			oldCmd := command_registry.Commands.FindCommand("fake-command")
			Expect(oldCmd).ToNot(Equal(updatedCmd))

			command_registry.Commands.SetCommand(updatedCmd)
			oldCmd = command_registry.Commands.FindCommand("fake-command")
			Expect(oldCmd).To(Equal(updatedCmd))
		})
	})

	Describe("Commands()", func() {
		Context("when there are commands registered", func() {
			BeforeEach(func() {
				command_registry.Register(FakeCommand1{})
				command_registry.Register(FakeCommand2{})
				command_registry.Register(FakeCommand3{})
			})

			AfterEach(func() {
				command_registry.Commands.RemoveCommand("fake-command1")
				command_registry.Commands.RemoveCommand("fake-command2")
				command_registry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
			})

			It("returns the total number of registered commands", func() {
				Expect(command_registry.Commands.TotalCommands()).To(Equal(3))
			})
		})

		It("returns 0 when there are no commands registered", func() {
			Expect(command_registry.Commands.TotalCommands()).To(Equal(0))
		})
	})

	Describe("Metadatas()", func() {
		BeforeEach(func() {
			command_registry.Register(FakeCommand1{})
			command_registry.Register(FakeCommand2{})
			command_registry.Register(FakeCommand3{})
		})

		AfterEach(func() {
			command_registry.Commands.RemoveCommand("fake-command")
			command_registry.Commands.RemoveCommand("fake-command2")
			command_registry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
		})

		It("returns all the metadata for all registered commands", func() {
			Expect(len(command_registry.Commands.Metadatas())).To(Equal(3))
		})
	})

	Describe("RemoveCommand()", func() {
		BeforeEach(func() {
			command_registry.Register(FakeCommand1{})
		})

		It("removes the command in registry with command name provided", func() {
			command_registry.Commands.RemoveCommand("fake-command")
			Expect(command_registry.Commands.CommandExists("fake-command")).To(BeFalse())
		})
	})

	Describe("MaxCommandNameLength()", func() {
		BeforeEach(func() {
			command_registry.Register(FakeCommand1{})
			command_registry.Register(FakeCommand2{})
			command_registry.Register(FakeCommand3{})
		})

		AfterEach(func() {
			command_registry.Commands.RemoveCommand("fake-command")
			command_registry.Commands.RemoveCommand("fake-command2")
			command_registry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
		})

		It("returns the length of the longest command name", func() {
			maxLen := command_registry.Commands.MaxCommandNameLength()
			Expect(maxLen).To(Equal(len("this-is-a-really-long-command-name-123123123123123123123")))
		})
	})

	Describe("CommandUsage()", func() {
		BeforeEach(func() {
			command_registry.Register(FakeCommand1{})
		})

		AfterEach(func() {
			command_registry.Commands.RemoveCommand("fake-command")
		})

		It("prints the name, description and usage of a command", func() {
			o := command_registry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"NAME:"},
				[]string{"   fake-command", "Description"},
				[]string{"USAGE:"},
			))
		})

		It("prints the flag options", func() {
			o := command_registry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"NAME:"},
				[]string{"USAGE:"},
				[]string{"OPTIONS:"},
				[]string{"intFlag", "Usage for"},
			))
		})

		It("replaces 'CF_NAME' with executable name from os.Arg[0]", func() {
			o := command_registry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"USAGE:"},
				[]string{"command_registry.test", "Usage of"},
			))
			Consistently(outputs).ShouldNot(ContainSubstrings([]string{"CF_NAME"}))
		})
	})
})
