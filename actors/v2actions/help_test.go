package v2actions_test

import (
	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Help Actions", func() {
	var actor Actor
	BeforeEach(func() {
		actor = NewActor()
	})

	Describe("CommandInfoByName", func() {
		Context("when the command exists", func() {
			Context("when passed the command name", func() {
				It("returns command info", func() {
					commandInfo, err := actor.CommandInfoByName(v2.Commands, "app")
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
					Expect(commandInfo.RelatedCommands).To(Equal([]string{
						"apps", "events", "logs", "map-route", "push", "unmap-route",
					}))
				})

				Context("when the command uses timeout environment variables", func() {
					DescribeTable("has timeout environment variables",
						func(setup func() (CommandInfo, error)) {
							commandInfo, err := setup()
							Expect(err).NotTo(HaveOccurred())

							Expect(commandInfo.Environment).To(ConsistOf(
								EnvironmentVariable{
									Name:         "CF_STAGING_TIMEOUT",
									Description:  "Max wait time for buildpack staging, in minutes",
									DefaultValue: "15",
								},
								EnvironmentVariable{
									Name:         "CF_STARTUP_TIMEOUT",
									Description:  "Max wait time for app instance startup, in minutes",
									DefaultValue: "5",
								}))
						},

						Entry("push command", func() (CommandInfo, error) {
							return actor.CommandInfoByName(v2.Commands, "push")
						}),

						Entry("start command", func() (CommandInfo, error) {
							return actor.CommandInfoByName(v2.Commands, "start")
						}),

						Entry("restart command", func() (CommandInfo, error) {
							return actor.CommandInfoByName(v2.Commands, "restart")
						}),

						Entry("restage command", func() (CommandInfo, error) {
							return actor.CommandInfoByName(v2.Commands, "restage")
						}),

						Entry("copy-source command", func() (CommandInfo, error) {
							return actor.CommandInfoByName(v2.Commands, "copy-source")
						}),
					)
				})

				Context("when the command does not use environment variables", func() {
					It("does not have environment variables", func() {
						commandInfo, err := actor.CommandInfoByName(v2.Commands, "app")
						Expect(err).NotTo(HaveOccurred())

						Expect(commandInfo.Environment).To(BeEmpty())
					})
				})
			})

			Context("when passed the command alias", func() {
				It("returns command info", func() {
					commandInfo, err := actor.CommandInfoByName(v2.Commands, "h")
					Expect(err).NotTo(HaveOccurred())

					Expect(commandInfo.Name).To(Equal("help"))
					Expect(commandInfo.Description).To(Equal("Show help"))
					Expect(commandInfo.Alias).To(Equal("h"))
					Expect(commandInfo.Usage).To(Equal("CF_NAME help [COMMAND]"))
					Expect(commandInfo.Flags).To(ConsistOf(
						CommandFlag{
							Short:       "a",
							Long:        "",
							Description: "All available CLI commands",
						},
					))
				})
			})
		})

		Context("when the command does not exist", func() {
			It("returns err", func() {
				_, err := actor.CommandInfoByName(v2.Commands, "does-not-exist")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ErrorInvalidCommand{CommandName: "does-not-exist"}))
			})
		})
	})

	Describe("CommandInfos", func() {
		It("returns back all the command's names and descriptions", func() {
			commands := actor.CommandInfos(v2.Commands)

			Expect(len(commands)).To(BeNumerically(">=", 153))
			Expect(commands["app"]).To(Equal(CommandInfo{
				Name:        "app",
				Description: "Display health and status for app",
			}))
			Expect(commands["curl"]).To(Equal(CommandInfo{
				Name:        "curl",
				Description: "Executes a request to the targeted API endpoint",
			}))
			Expect(commands["apps"]).To(Equal(CommandInfo{
				Name:        "apps",
				Description: "List all apps in the target space",
				Alias:       "a",
			}))
			Expect(commands["routes"]).To(Equal(CommandInfo{
				Name:        "routes",
				Description: "List all routes in the current space or the current organization",
				Alias:       "r",
			}))
		})
	})
})
