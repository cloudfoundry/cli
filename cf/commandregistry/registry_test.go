package commandregistry_test

import (
	"strings"

	"code.cloudfoundry.org/cli/cf/commandregistry"

	. "code.cloudfoundry.org/cli/cf/commandregistry/fakecommand"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "code.cloudfoundry.org/cli/cf/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRegistry", func() {
	BeforeEach(func() {
		commandregistry.Commands = commandregistry.NewRegistry() // because other tests load all the commands into the registry
	})

	Context("i18n", func() {
		It("initialize i18n T() func", func() {
			Expect(T).ToNot(BeNil())
		})
	})

	Describe("Register()", func() {
		AfterEach(func() {
			commandregistry.Commands.RemoveCommand("fake-command2")
		})

		It("registers a command and its alias into the Command Registry map", func() {
			Expect(commandregistry.Commands.CommandExists("fake-command2")).To(BeFalse())
			Expect(commandregistry.Commands.CommandExists("fc2")).To(BeFalse())

			commandregistry.Register(FakeCommand2{})

			Expect(commandregistry.Commands.CommandExists("fake-command2")).To(BeTrue())
			Expect(commandregistry.Commands.CommandExists("fc2")).To(BeTrue())
		})
	})

	Describe("CommandExists()", func() {
		Context("when the command has been registered", func() {
			BeforeEach(func() {
				commandregistry.Register(FakeCommand1{})
			})

			AfterEach(func() {
				commandregistry.Commands.RemoveCommand("fake-command")
			})

			It("returns true the command exists in the list", func() {
				Expect(commandregistry.Commands.CommandExists("fake-command")).To(BeTrue())
			})

			It("returns true if the alias exists", func() {
				Expect(commandregistry.Commands.CommandExists("fc1")).To(BeTrue())
			})
		})

		It("returns false when the command has not been registered", func() {
			Expect(commandregistry.Commands.CommandExists("non-exist-cmd")).To(BeFalse())
		})

		It("returns false if the command name is an empty string", func() {
			Expect(commandregistry.Commands.CommandExists("")).To(BeFalse())
		})
	})

	Describe("FindCommand()", func() {
		Context("when the command has been registered", func() {
			BeforeEach(func() {
				commandregistry.Register(FakeCommand1{})
			})

			AfterEach(func() {
				commandregistry.Commands.RemoveCommand("fake-command")
			})

			It("returns the command when the command's name is given", func() {
				cmd := commandregistry.Commands.FindCommand("fake-command")
				Expect(cmd.MetaData().Usage[0]).To(ContainSubstring("Usage of fake-command"))
				Expect(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})

			It("returns the command when the command's alias is given", func() {
				cmd := commandregistry.Commands.FindCommand("fc1")
				Expect(cmd.MetaData().Usage[0]).To(ContainSubstring("Usage of fake-command"))
				Expect(cmd.MetaData().Description).To(Equal("Description for fake-command"))
			})
		})

		It("returns nil when the command has not been registered", func() {
			cmd := commandregistry.Commands.FindCommand("fake-command")
			Expect(cmd).To(BeNil())
		})
	})

	Describe("ShowAllCommands()", func() {
		BeforeEach(func() {
			commandregistry.Register(FakeCommand1{})
			commandregistry.Register(FakeCommand2{})
			commandregistry.Register(FakeCommand3{})
		})

		AfterEach(func() {
			commandregistry.Commands.RemoveCommand("fake-command")
			commandregistry.Commands.RemoveCommand("fake-command2")
			commandregistry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
		})

		It("show all the commands in registry", func() {
			cmds := commandregistry.Commands.ListCommands()
			Expect(cmds).To(ContainElement("fake-command2"))
			Expect(cmds).To(ContainElement("this-is-a-really-long-command-name-123123123123123123123"))
			Expect(cmds).To(ContainElement("fake-command"))
		})
	})

	Describe("SetCommand()", func() {
		It("replaces the command in registry with command provided", func() {
			updatedCmd := FakeCommand1{Data: "This is new data"}
			oldCmd := commandregistry.Commands.FindCommand("fake-command")
			Expect(oldCmd).ToNot(Equal(updatedCmd))

			commandregistry.Commands.SetCommand(updatedCmd)
			oldCmd = commandregistry.Commands.FindCommand("fake-command")
			Expect(oldCmd).To(Equal(updatedCmd))
		})
	})

	Describe("TotalCommands()", func() {
		Context("when there are commands registered", func() {
			BeforeEach(func() {
				commandregistry.Register(FakeCommand1{})
				commandregistry.Register(FakeCommand2{})
				commandregistry.Register(FakeCommand3{})
			})

			AfterEach(func() {
				commandregistry.Commands.RemoveCommand("fake-command")
				commandregistry.Commands.RemoveCommand("fake-command2")
				commandregistry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
			})

			It("returns the total number of registered commands", func() {
				Expect(commandregistry.Commands.TotalCommands()).To(Equal(3))
			})
		})

		It("returns 0 when there are no commands registered", func() {
			Expect(commandregistry.Commands.TotalCommands()).To(Equal(0))
		})
	})

	Describe("Metadatas()", func() {
		BeforeEach(func() {
			commandregistry.Register(FakeCommand1{})
			commandregistry.Register(FakeCommand2{})
			commandregistry.Register(FakeCommand3{})
		})

		AfterEach(func() {
			commandregistry.Commands.RemoveCommand("fake-command")
			commandregistry.Commands.RemoveCommand("fake-command2")
			commandregistry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
		})

		It("returns all the metadata for all registered commands", func() {
			Expect(len(commandregistry.Commands.Metadatas())).To(Equal(3))
		})
	})

	Describe("RemoveCommand()", func() {
		BeforeEach(func() {
			commandregistry.Register(FakeCommand1{})
		})

		It("removes the command in registry with command name provided", func() {
			commandregistry.Commands.RemoveCommand("fake-command")
			Expect(commandregistry.Commands.CommandExists("fake-command")).To(BeFalse())
		})
	})

	Describe("MaxCommandNameLength()", func() {
		BeforeEach(func() {
			commandregistry.Register(FakeCommand1{})
			commandregistry.Register(FakeCommand2{})
			commandregistry.Register(FakeCommand3{})
		})

		AfterEach(func() {
			commandregistry.Commands.RemoveCommand("fake-command")
			commandregistry.Commands.RemoveCommand("fake-command2")
			commandregistry.Commands.RemoveCommand("this-is-a-really-long-command-name-123123123123123123123") // fake-command3
		})

		It("returns the length of the longest command name", func() {
			maxLen := commandregistry.Commands.MaxCommandNameLength()
			Expect(maxLen).To(Equal(len("this-is-a-really-long-command-name-123123123123123123123")))
		})
	})

	Describe("CommandUsage()", func() {
		BeforeEach(func() {
			commandregistry.Register(FakeCommand1{})
		})

		AfterEach(func() {
			commandregistry.Commands.RemoveCommand("fake-command")
		})

		It("prints the name, description and usage of a command", func() {
			o := commandregistry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"NAME:"},
				[]string{"   fake-command", "Description"},
				[]string{"USAGE:"},
			))
		})

		Context("i18n translations", func() {
			var originalT func(string, ...interface{}) string

			BeforeEach(func() {
				originalT = T
			})

			AfterEach(func() {
				T = originalT
			})

			It("includes ':' in caption translation strings for language like French to be translated correctly", func() {
				nameCaption := "NAME:"
				aliasCaption := "ALIAS:"
				usageCaption := "USAGE:"
				optionsCaption := "OPTIONS:"
				captionCheckCount := 0

				T = func(translationID string, args ...interface{}) string {
					if strings.HasPrefix(translationID, "NAME") {
						Expect(translationID).To(Equal(nameCaption))
						captionCheckCount += 1
					} else if strings.HasPrefix(translationID, "ALIAS") {
						Expect(translationID).To(Equal(aliasCaption))
						captionCheckCount += 1
					} else if strings.HasPrefix(translationID, "USAGE") {
						Expect(translationID).To(Equal(usageCaption))
						captionCheckCount += 1
					} else if strings.HasPrefix(translationID, "OPTIONS") {
						Expect(translationID).To(Equal(optionsCaption))
						captionCheckCount += 1
					}

					return translationID
				}

				commandregistry.Commands.CommandUsage("fake-command")
			})
		})

		It("prints the flag options", func() {
			o := commandregistry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"NAME:"},
				[]string{"USAGE:"},
				[]string{"OPTIONS:"},
				[]string{"intFlag", "Usage for"},
			))
		})

		It("replaces 'CF_NAME' with executable name from os.Arg[0]", func() {
			o := commandregistry.Commands.CommandUsage("fake-command")
			outputs := strings.Split(o, "\n")
			Expect(outputs).To(BeInDisplayOrder(
				[]string{"USAGE:"},
				[]string{"cf", "Usage of"},
			))
			Consistently(outputs).ShouldNot(ContainSubstrings([]string{"CF_NAME"}))
		})
	})
})
