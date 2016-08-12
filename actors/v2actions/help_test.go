package v2actions_test

import (
	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Help Actions", func() {
	var actor Actor
	BeforeEach(func() {
		actor = NewActor()
	})

	Describe("GetCommandInfo", func() {
		Context("when the command exists", func() {
			Context("when passed the command name", func() {
				It("returns command info", func() {
					commandInfo, err := actor.GetCommandInfo(v2.Commands, "app")
					Expect(err).NotTo(HaveOccurred())

					Expect(commandInfo.Name).To(Equal("app"))
					Expect(commandInfo.Description).To(Equal("Display health and status for app"))
					Expect(commandInfo.Alias).To(BeEmpty())
					Expect(commandInfo.Usage).To(Equal("CF_NAME app APP_NAME"))
					Expect(commandInfo.Flags).To(HaveLen(1))
					Expect(commandInfo.Flags).To(ContainElement(CommandFlag{
						Short:       "",
						Long:        "guid",
						Description: "Retrieve and display the given app's guid.  All other health and status output for the app is suppressed.",
					}))
				})
			})

			Context("when passed the command alias", func() {
				It("returns command info", func() {
					commandInfo, err := actor.GetCommandInfo(v2.Commands, "h")
					Expect(err).NotTo(HaveOccurred())

					Expect(commandInfo.Name).To(Equal("help"))
					Expect(commandInfo.Description).To(Equal("Show help"))
					Expect(commandInfo.Alias).To(Equal("h"))
					Expect(commandInfo.Usage).To(Equal("CF_NAME help [COMMAND]"))
					Expect(commandInfo.Flags).To(BeEmpty())
				})
			})
		})

		Context("when the command does not exist", func() {
			It("returns err", func() {
				_, err := actor.GetCommandInfo(v2.Commands, "does-not-exist")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ErrorInvalidCommand{CommandName: "does-not-exist"}))
			})
		})
	})

	Describe("GetAllNamesAndDescriptions", func() {
		It("returns back all the command's names and descriptions", func() {
			commands := actor.GetAllNamesAndDescriptions(v2.Commands)

			Expect(len(commands)).To(BeNumerically(">=", 153))
			Expect(commands["app"]).To(Equal(CommandInfo{
				Name:        "app",
				Description: "Display health and status for app",
			}))
			Expect(commands["curl"]).To(Equal(CommandInfo{
				Name:        "curl",
				Description: "Executes a request to the targeted API endpoint",
			}))
		})
	})
})
