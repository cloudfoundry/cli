package v2_test

import (
	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	"code.cloudfoundry.org/cli/commands/flags"
	. "code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Help Command", func() {
	var (
		fakeUI     *ui.UI
		fakeActor  *v2fakes.FakeHelpActor
		cmd        HelpCommand
		fakeConfig *commandsfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeActor = new(v2fakes.FakeHelpActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("faceman")

		cmd = HelpCommand{
			UI:     fakeUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	Context("providing help for a specific command", func() {
		Describe("built-in command", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flags.CommandName{
					CommandName: "help",
				}

				commandInfo := v2actions.CommandInfo{
					Name:        "help",
					Description: "Show help",
					Usage:       "CF_NAME help [COMMAND]",
					Alias:       "h",
				}
				fakeActor.CommandInfoByNameReturns(commandInfo, nil)
			})

			It("displays the name for help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("NAME:"))
				Expect(fakeUI.Out).To(Say("   help - Show help"))

				Expect(fakeActor.CommandInfoByNameCallCount()).To(Equal(1))
				_, commandName := fakeActor.CommandInfoByNameArgsForCall(0)
				Expect(commandName).To(Equal("help"))
			})

			It("displays the usage for help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("NAME:"))
				Expect(fakeUI.Out).To(Say("USAGE:"))
				Expect(fakeUI.Out).To(Say("   faceman help \\[COMMAND\\]"))
			})

			Describe("related commands", func() {
				Context("when the command has related commands", func() {
					BeforeEach(func() {
						commandInfo := v2actions.CommandInfo{
							Name:            "app",
							RelatedCommands: []string{"broccoli", "tomato"},
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					It("displays the related commands for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("NAME:"))
						Expect(fakeUI.Out).To(Say("SEE ALSO:"))
						Expect(fakeUI.Out).To(Say("   broccoli, tomato"))
					})
				})

				Context("when the command does not have related commands", func() {
					It("displays the related commands for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("NAME:"))
						Expect(fakeUI.Out).NotTo(Say("SEE ALSO:"))
					})
				})
			})

			Describe("aliases", func() {
				Context("when the command has an alias", func() {
					It("displays the alias for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).To(Say("USAGE:"))
						Expect(fakeUI.Out).To(Say("ALIAS:"))
						Expect(fakeUI.Out).To(Say("   h"))
					})
				})

				Context("when the command does not have an alias", func() {
					BeforeEach(func() {
						cmd.OptionalArgs = flags.CommandName{
							CommandName: "app",
						}

						commandInfo := v2actions.CommandInfo{
							Name: "app",
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					It("no alias is displayed", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeUI.Out).ToNot(Say("ALIAS:"))
					})
				})
			})

			Describe("options", func() {
				Context("when the command has options", func() {
					BeforeEach(func() {
						cmd.OptionalArgs = flags.CommandName{
							CommandName: "push",
						}
						commandInfo := v2actions.CommandInfo{
							Name: "push",
							Flags: []v2actions.CommandFlag{
								{
									Long:        "no-hostname",
									Description: "Map the root domain to this app",
								},
								{
									Short:       "b",
									Description: "Custom buildpack by name (e.g. my-buildpack) or Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'",
								},
								{
									Long:        "hostname",
									Short:       "n",
									Description: "Hostname (e.g. my-subdomain)",
								},
							},
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					Context("only has a long option", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeUI.Out).To(Say("USAGE:"))
							Expect(fakeUI.Out).To(Say("OPTIONS:"))
							Expect(fakeUI.Out).To(Say("--no-hostname\\s+Map the root domain to this app"))
						})
					})

					Context("only has a short option", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeUI.Out).To(Say("USAGE:"))
							Expect(fakeUI.Out).To(Say("OPTIONS:"))
							Expect(fakeUI.Out).To(Say("-b\\s+Custom buildpack by name \\(e.g. my-buildpack\\) or Git URL \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git'\\) or Git URL with a branch or tag \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag\\). To use built-in buildpacks only, specify 'default' or 'null'"))
						})
					})

					Context("has long and short options", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeUI.Out).To(Say("USAGE:"))
							Expect(fakeUI.Out).To(Say("OPTIONS:"))
							Expect(fakeUI.Out).To(Say("--hostname, -n\\s+Hostname \\(e.g. my-subdomain\\)"))
						})
					})

					Context("has hidden options", func() {
						It("does not display the hidden option", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(fakeUI.Out).ToNot(Say("--app-ports"))
						})
					})
				})
			})
		})

		Describe("Environment", func() {
			Context("has environment variables", func() {
				var envVars []v2actions.EnvironmentVariable

				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "push",
					}
					envVars = []v2actions.EnvironmentVariable{
						v2actions.EnvironmentVariable{
							Name:         "CF_STAGING_TIMEOUT",
							Description:  "Max wait time for buildpack staging, in minutes",
							DefaultValue: "15",
						},
						v2actions.EnvironmentVariable{
							Name:         "CF_STARTUP_TIMEOUT",
							Description:  "Max wait time for app instance startup, in minutes",
							DefaultValue: "5",
						},
					}
					commandInfo := v2actions.CommandInfo{
						Name:        "push",
						Environment: envVars,
					}

					fakeActor.CommandInfoByNameReturns(commandInfo, nil)
				})

				It("displays the timeouts under environment", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("ENVIRONMENT:"))
					Expect(fakeUI.Out).To(Say(`
   CF_STAGING_TIMEOUT=15        Max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5         Max wait time for app instance startup, in minutes
`))
				})
			})

			Context("does not have any associated environment variables", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flags.CommandName{
						CommandName: "app",
					}
					commandInfo := v2actions.CommandInfo{
						Name: "app",
					}

					fakeActor.CommandInfoByNameReturns(commandInfo, nil)
				})

				It("does not show the environment section", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeUI.Out).ToNot(Say("ENVIRONMENT:"))
				})
			})
		})

		Describe("plug-in command", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flags.CommandName{
					CommandName: "enable-diego",
				}

				fakeConfig.PluginsReturns(map[string]configv3.Plugin{
					"Diego-Enabler": configv3.Plugin{
						Commands: []configv3.PluginCommand{
							{
								Name:     "enable-diego",
								Alias:    "ed",
								HelpText: "enable Diego support for an app",
								UsageDetails: configv3.PluginUsageDetails{
									Usage: "faceman diego-enabler this and that and a little stuff",
									Options: map[string]string{
										"--first":        "foobar",
										"--second-third": "baz",
									},
								},
							},
						},
					},
				})

				fakeActor.CommandInfoByNameReturns(v2actions.CommandInfo{},
					v2actions.ErrorInvalidCommand{CommandName: "enable-diego"})
			})

			It("displays the plugin's help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("enable-diego - enable Diego support for an app"))
				Expect(fakeUI.Out).To(Say("faceman diego-enabler this and that and a little stuff"))
				Expect(fakeUI.Out).To(Say("ALIAS:"))
				Expect(fakeUI.Out).To(Say("ed"))
				Expect(fakeUI.Out).To(Say("--first\\s+foobar"))
				Expect(fakeUI.Out).To(Say("--second-third\\s+baz"))
			})
		})
	})

	Describe("help for common commands", func() {
		BeforeEach(func() {
			cmd.OptionalArgs = flags.CommandName{
				CommandName: "",
			}
			cmd.AllCommands = false
			cmd.Actor = v2actions.NewActor(nil)
		})

		It("returns a list of only the common commands", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeUI.Out).To(Say("faceman version BUILT_FROM_SOURCE-BUILT_AT_UNKNOWN_TIME, Cloud Foundry command line tool"))
			Expect(fakeUI.Out).To(Say("Usage: faceman \\[global options\\] command \\[arguments...\\] \\[command options\\]"))

			Expect(fakeUI.Out).To(Say("Before getting started:"))
			Expect(fakeUI.Out).To(Say("help,h\\s+logout,lo"))

			Expect(fakeUI.Out).To(Say("Application lifecycle:"))
			Expect(fakeUI.Out).To(Say("apps,a\\s+logs\\s+set-env,se"))

			Expect(fakeUI.Out).To(Say("Services integration:"))
			Expect(fakeUI.Out).To(Say("marketplace,m\\s+create-user-provided-service,cups"))
			Expect(fakeUI.Out).To(Say("services,s\\s+update-user-provided-service,uups"))

			Expect(fakeUI.Out).To(Say("Route and domain management:"))
			Expect(fakeUI.Out).To(Say("routes,r\\s+delete-route\\s+create-domain"))
			Expect(fakeUI.Out).To(Say("domains\\s+map-route"))

			Expect(fakeUI.Out).To(Say("Space management:"))
			Expect(fakeUI.Out).To(Say("spaces\\s+create-space\\s+set-space-role"))

			Expect(fakeUI.Out).To(Say("Org management:"))
			Expect(fakeUI.Out).To(Say("orgs,o\\s+set-org-role"))

			Expect(fakeUI.Out).To(Say("CLI plugin management:"))
			Expect(fakeUI.Out).To(Say("plugins\\s+add-plugin-repo\\s+repo-plugins"))

			Expect(fakeUI.Out).To(Say("Global options:"))
			Expect(fakeUI.Out).To(Say("--help, -h\\s+Show help"))

			Expect(fakeUI.Out).To(Say("'cf help -a' lists all commands with short descriptions. See 'cf help <command>'"))
		})

		Context("when there are multiple installed plugins", func() {
			BeforeEach(func() {
				fakeConfig.PluginsReturns(map[string]configv3.Plugin{
					"some-plugin": configv3.Plugin{
						Commands: []configv3.PluginCommand{
							{
								Name:     "enable",
								HelpText: "enable command",
							},
							{
								Name:     "disable",
								HelpText: "disable command",
							},
							{
								Name:     "some-other-command",
								HelpText: "does something",
							},
						},
					},
					"Some-other-plugin": configv3.Plugin{
						Commands: []configv3.PluginCommand{
							{
								Name:     "some-other-plugin-command",
								HelpText: "does some other thing",
							},
						},
					},
					"the-last-plugin": configv3.Plugin{
						Commands: []configv3.PluginCommand{
							{
								Name:     "last-plugin-command",
								HelpText: "does the last thing",
							},
						},
					},
				})
			})

			It("returns the plugin commands organized by plugin and sorted in alphabetical order", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("Commands offered by installed plugins:"))
				Expect(fakeUI.Out).To(Say("some-other-plugin-command\\s+enable\\s+last-plugin-command"))
				Expect(fakeUI.Out).To(Say("disable\\s+some-other-command"))

			})
		})
	})

	Describe("providing help for all commands", func() {
		Context("when a command is not provided", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flags.CommandName{
					CommandName: "",
				}
				cmd.AllCommands = true

				cmd.Actor = v2actions.NewActor(nil)
				fakeConfig.PluginsReturns(map[string]configv3.Plugin{
					"Diego-Enabler": configv3.Plugin{
						Commands: []configv3.PluginCommand{
							{
								Name:     "enable-diego",
								HelpText: "enable Diego support for an app",
							},
						},
					},
				})
			})

			It("returns a list of all commands", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeUI.Out).To(Say("NAME:"))
				Expect(fakeUI.Out).To(Say("faceman - A command line tool to interact with Cloud Foundry"))
				Expect(fakeUI.Out).To(Say("USAGE:"))
				Expect(fakeUI.Out).To(Say("faceman \\[global options\\] command \\[arguments...\\] \\[command options\\]"))
				Expect(fakeUI.Out).To(Say("VERSION:"))
				Expect(fakeUI.Out).To(Say("BUILT_FROM_SOURCE-BUILT_AT_UNKNOWN_TIME"))

				Expect(fakeUI.Out).To(Say("GETTING STARTED:"))
				Expect(fakeUI.Out).To(Say("help\\s+Show help"))
				Expect(fakeUI.Out).To(Say("api\\s+Set or view target api url"))

				Expect(fakeUI.Out).To(Say("APPS:"))
				Expect(fakeUI.Out).To(Say("apps\\s+List all apps in the target space"))
				Expect(fakeUI.Out).To(Say("ssh-enabled\\s+Reports whether SSH is enabled on an application container instance"))

				Expect(fakeUI.Out).To(Say("SERVICES:"))
				Expect(fakeUI.Out).To(Say("marketplace\\s+List available offerings in the marketplace"))
				Expect(fakeUI.Out).To(Say("create-service\\s+Create a service instance"))

				Expect(fakeUI.Out).To(Say("ORGS:"))
				Expect(fakeUI.Out).To(Say("orgs\\s+List all orgs"))
				Expect(fakeUI.Out).To(Say("delete-org\\s+Delete an org"))

				Expect(fakeUI.Out).To(Say("SPACES:"))
				Expect(fakeUI.Out).To(Say("spaces\\s+List all spaces in an org"))
				Expect(fakeUI.Out).To(Say("allow-space-ssh\\s+Allow SSH access for the space"))

				Expect(fakeUI.Out).To(Say("DOMAINS:"))
				Expect(fakeUI.Out).To(Say("domains\\s+List domains in the target org"))
				Expect(fakeUI.Out).To(Say("router-groups\\s+List router groups"))

				Expect(fakeUI.Out).To(Say("ROUTES:"))
				Expect(fakeUI.Out).To(Say("routes\\s+List all routes in the current space or the current organization"))
				Expect(fakeUI.Out).To(Say("unmap-route\\s+Remove a url route from an app"))

				Expect(fakeUI.Out).To(Say("BUILDPACKS:"))
				Expect(fakeUI.Out).To(Say("buildpacks\\s+List all buildpacks"))
				Expect(fakeUI.Out).To(Say("delete-buildpack\\s+Delete a buildpack"))

				Expect(fakeUI.Out).To(Say("USER ADMIN:"))
				Expect(fakeUI.Out).To(Say("create-user\\s+Create a new user"))
				Expect(fakeUI.Out).To(Say("space-users\\s+Show space users by role"))

				Expect(fakeUI.Out).To(Say("ORG ADMIN:"))
				Expect(fakeUI.Out).To(Say("quotas\\s+List available usage quotas"))
				Expect(fakeUI.Out).To(Say("delete-quota\\s+Delete a quota"))

				Expect(fakeUI.Out).To(Say("SPACE ADMIN:"))
				Expect(fakeUI.Out).To(Say("space-quotas\\s+List available space resource quotas"))
				Expect(fakeUI.Out).To(Say("set-space-quota\\s+Assign a space quota definition to a space"))

				Expect(fakeUI.Out).To(Say("SERVICE ADMIN:"))
				Expect(fakeUI.Out).To(Say("service-auth-tokens\\s+List service auth tokens"))
				Expect(fakeUI.Out).To(Say("service-access\\s+List service access settings"))

				Expect(fakeUI.Out).To(Say("SECURITY GROUP:"))
				Expect(fakeUI.Out).To(Say("security-group\\s+Show a single security group"))
				Expect(fakeUI.Out).To(Say("staging-security-groups\\s+List security groups in the staging set for applications"))

				Expect(fakeUI.Out).To(Say("ENVIRONMENT VARIABLE GROUPS:"))
				Expect(fakeUI.Out).To(Say("running-environment-variable-group\\s+Retrieve the contents of the running environment variable group"))
				Expect(fakeUI.Out).To(Say("set-running-environment-variable-group\\s+Pass parameters as JSON to create a running environment variable group"))

				Expect(fakeUI.Out).To(Say("FEATURE FLAGS:"))
				Expect(fakeUI.Out).To(Say("feature-flags\\s+Retrieve list of feature flags with status of each flag-able feature"))
				Expect(fakeUI.Out).To(Say("disable-feature-flag\\s+Disable the use of a feature so that users have access to and can use the feature"))

				Expect(fakeUI.Out).To(Say("ADVANCED:"))
				Expect(fakeUI.Out).To(Say("curl\\s+Executes a request to the targeted API endpoint"))
				Expect(fakeUI.Out).To(Say("ssh-code\\s+Get a one time password for ssh clients"))

				Expect(fakeUI.Out).To(Say("ADD/REMOVE PLUGIN REPOSITORY:"))
				Expect(fakeUI.Out).To(Say("add-plugin-repo\\s+Add a new plugin repository"))
				Expect(fakeUI.Out).To(Say("repo-plugins\\s+List all available plugins in specified repository or in all added repositories"))

				Expect(fakeUI.Out).To(Say("ADD/REMOVE PLUGIN:"))
				Expect(fakeUI.Out).To(Say("plugins\\s+List all available plugin commands"))
				Expect(fakeUI.Out).To(Say("uninstall-plugin\\s+Uninstall the plugin defined in command argument"))

				Expect(fakeUI.Out).To(Say("INSTALLED PLUGIN COMMANDS:"))
				Expect(fakeUI.Out).To(Say("enable-diego\\s+enable Diego support for an app"))

				Expect(fakeUI.Out).To(Say("ENVIRONMENT VARIABLES:"))
				Expect(fakeUI.Out).To(Say("CF_COLOR=false\\s+Do not colorize output"))
				Expect(fakeUI.Out).To(Say("CF_DIAL_TIMEOUT=5\\s+Max wait time to establish a connection, including name resolution, in seconds"))

				Expect(fakeUI.Out).To(Say("GLOBAL OPTIONS:"))
				Expect(fakeUI.Out).To(Say("--help, -h\\s+Show help"))
			})

			Context("when there are multiple installed plugins", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns(map[string]configv3.Plugin{
						"some-plugin": configv3.Plugin{
							Commands: []configv3.PluginCommand{
								{
									Name:     "enable",
									HelpText: "enable command",
								},
								{
									Name:     "disable",
									HelpText: "disable command",
								},
								{
									Name:     "some-other-command",
									HelpText: "does something",
								},
							},
						},
						"Some-other-plugin": configv3.Plugin{
							Commands: []configv3.PluginCommand{
								{
									Name:     "some-other-plugin-command",
									HelpText: "does some other thing",
								},
							},
						},
						"the-last-plugin": configv3.Plugin{
							Commands: []configv3.PluginCommand{
								{
									Name:     "last-plugin-command",
									HelpText: "does the last thing",
								},
							},
						},
					})
				})

				It("returns the plugin commands organized by plugin and sorted in alphabetical order", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say(`INSTALLED PLUGIN COMMANDS:.*
\s+some-other-plugin-command\s+does some other thing.*
\s+disable\s+disable command.*
\s+enable\s+enable command.*
\s+some-other-command\s+does something.*
\s+last-plugin-command\s+does the last thing`))
				})
			})
		})
	})
})
