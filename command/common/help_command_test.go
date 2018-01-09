package common_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/common/commonfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("help Command", func() {
	var (
		testUI     *ui.UI
		fakeActor  *commonfakes.FakeHelpActor
		cmd        HelpCommand
		fakeConfig *commandfakes.FakeConfig
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeActor = new(commonfakes.FakeHelpActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("faceman")
		fakeConfig.BinaryVersionReturns("face2.0-yesterday")

		cmd = HelpCommand{
			UI:     testUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	Context("providing help for a specific command", func() {
		Describe("built-in command", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flag.CommandName{
					CommandName: "help",
				}

				commandInfo := sharedaction.CommandInfo{
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

				Expect(testUI.Out).To(Say("NAME:"))
				Expect(testUI.Out).To(Say("   help - Show help"))

				Expect(fakeActor.CommandInfoByNameCallCount()).To(Equal(1))
				_, commandName := fakeActor.CommandInfoByNameArgsForCall(0)
				Expect(commandName).To(Equal("help"))
			})

			It("displays the usage for help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("NAME:"))
				Expect(testUI.Out).To(Say("USAGE:"))
				Expect(testUI.Out).To(Say("   faceman help \\[COMMAND\\]"))
			})

			Describe("related commands", func() {
				Context("when the command has related commands", func() {
					BeforeEach(func() {
						commandInfo := sharedaction.CommandInfo{
							Name:            "app",
							RelatedCommands: []string{"broccoli", "tomato"},
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					It("displays the related commands for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("NAME:"))
						Expect(testUI.Out).To(Say("SEE ALSO:"))
						Expect(testUI.Out).To(Say("   broccoli, tomato"))
					})
				})

				Context("when the command does not have related commands", func() {
					It("displays the related commands for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("NAME:"))
						Expect(testUI.Out).NotTo(Say("SEE ALSO:"))
					})
				})
			})

			Describe("aliases", func() {
				Context("when the command has an alias", func() {
					It("displays the alias for help", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("USAGE:"))
						Expect(testUI.Out).To(Say("ALIAS:"))
						Expect(testUI.Out).To(Say("   h"))
					})
				})

				Context("when the command does not have an alias", func() {
					BeforeEach(func() {
						cmd.OptionalArgs = flag.CommandName{
							CommandName: "app",
						}

						commandInfo := sharedaction.CommandInfo{
							Name: "app",
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					It("no alias is displayed", func() {
						err := cmd.Execute(nil)
						Expect(err).ToNot(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("ALIAS:"))
					})
				})
			})

			Describe("options", func() {
				Context("when the command has options", func() {
					BeforeEach(func() {
						cmd.OptionalArgs = flag.CommandName{
							CommandName: "push",
						}
						commandInfo := sharedaction.CommandInfo{
							Name: "push",
							Flags: []sharedaction.CommandFlag{
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
								{
									Long:        "force",
									Short:       "f",
									Description: "do it",
									Default:     "yes",
								},
							},
						}
						fakeActor.CommandInfoByNameReturns(commandInfo, nil)
					})

					Context("only has a long option", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("USAGE:"))
							Expect(testUI.Out).To(Say("OPTIONS:"))
							Expect(testUI.Out).To(Say("--no-hostname\\s+Map the root domain to this app"))
						})
					})

					Context("only has a short option", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("USAGE:"))
							Expect(testUI.Out).To(Say("OPTIONS:"))
							Expect(testUI.Out).To(Say("-b\\s+Custom buildpack by name \\(e.g. my-buildpack\\) or Git URL \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git'\\) or Git URL with a branch or tag \\(e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag\\). To use built-in buildpacks only, specify 'default' or 'null'"))
						})
					})

					Context("has long and short options", func() {
						It("displays the options for app", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("USAGE:"))
							Expect(testUI.Out).To(Say("OPTIONS:"))
							Expect(testUI.Out).To(Say("--hostname, -n\\s+Hostname \\(e.g. my-subdomain\\)"))
						})
					})

					Context("has hidden options", func() {
						It("does not display the hidden option", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(testUI.Out).ToNot(Say("--app-ports"))
						})
					})

					Context("has a default for an option", func() {
						It("displays the default", func() {
							err := cmd.Execute(nil)
							Expect(err).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("do it \\(Default: yes\\)"))
						})
					})
				})
			})
		})

		Describe("Environment", func() {
			Context("has environment variables", func() {
				var envVars []sharedaction.EnvironmentVariable

				BeforeEach(func() {
					cmd.OptionalArgs = flag.CommandName{
						CommandName: "push",
					}
					envVars = []sharedaction.EnvironmentVariable{
						sharedaction.EnvironmentVariable{
							Name:         "CF_STAGING_TIMEOUT",
							Description:  "Max wait time for buildpack staging, in minutes",
							DefaultValue: "15",
						},
						sharedaction.EnvironmentVariable{
							Name:         "CF_STARTUP_TIMEOUT",
							Description:  "Max wait time for app instance startup, in minutes",
							DefaultValue: "5",
						},
					}
					commandInfo := sharedaction.CommandInfo{
						Name:        "push",
						Environment: envVars,
					}

					fakeActor.CommandInfoByNameReturns(commandInfo, nil)
				})

				It("displays the timeouts under environment", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("ENVIRONMENT:"))
					Expect(testUI.Out).To(Say(`
   CF_STAGING_TIMEOUT=15        Max wait time for buildpack staging, in minutes
   CF_STARTUP_TIMEOUT=5         Max wait time for app instance startup, in minutes
`))
				})
			})

			Context("does not have any associated environment variables", func() {
				BeforeEach(func() {
					cmd.OptionalArgs = flag.CommandName{
						CommandName: "app",
					}
					commandInfo := sharedaction.CommandInfo{
						Name: "app",
					}

					fakeActor.CommandInfoByNameReturns(commandInfo, nil)
				})

				It("does not show the environment section", func() {
					err := cmd.Execute(nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(testUI.Out).ToNot(Say("ENVIRONMENT:"))
				})
			})
		})

		Describe("plug-in command", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flag.CommandName{
					CommandName: "enable-diego",
				}

				fakeConfig.PluginsReturns([]configv3.Plugin{
					{Name: "Diego-Enabler",
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

				fakeActor.CommandInfoByNameReturns(sharedaction.CommandInfo{},
					actionerror.InvalidCommandError{CommandName: "enable-diego"})
			})

			It("displays the plugin's help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("enable-diego - enable Diego support for an app"))
				Expect(testUI.Out).To(Say("faceman diego-enabler this and that and a little stuff"))
				Expect(testUI.Out).To(Say("ALIAS:"))
				Expect(testUI.Out).To(Say("ed"))
				Expect(testUI.Out).To(Say("--first\\s+foobar"))
				Expect(testUI.Out).To(Say("--second-third\\s+baz"))
			})
		})

		Describe("plug-in alias", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flag.CommandName{
					CommandName: "ed",
				}

				fakeConfig.PluginsReturns([]configv3.Plugin{
					{
						Name: "Diego-Enabler",
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

				fakeActor.CommandInfoByNameReturns(sharedaction.CommandInfo{},
					actionerror.InvalidCommandError{CommandName: "enable-diego"})
			})

			It("displays the plugin's help", func() {
				err := cmd.Execute(nil)
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("enable-diego - enable Diego support for an app"))
				Expect(testUI.Out).To(Say("faceman diego-enabler this and that and a little stuff"))
				Expect(testUI.Out).To(Say("ALIAS:"))
				Expect(testUI.Out).To(Say("ed"))
				Expect(testUI.Out).To(Say("--first\\s+foobar"))
				Expect(testUI.Out).To(Say("--second-third\\s+baz"))
			})
		})
	})

	Describe("help for common commands", func() {
		BeforeEach(func() {
			cmd.OptionalArgs = flag.CommandName{
				CommandName: "",
			}
			cmd.AllCommands = false
			cmd.Actor = sharedaction.NewActor(nil)
		})

		It("returns a list of only the common commands", func() {
			err := cmd.Execute(nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("faceman version face2.0-yesterday, Cloud Foundry command line tool"))
			Expect(testUI.Out).To(Say("Usage: faceman \\[global options\\] command \\[arguments...\\] \\[command options\\]"))

			Expect(testUI.Out).To(Say("Before getting started:"))
			Expect(testUI.Out).To(Say("  help,h    logout,lo"))

			Expect(testUI.Out).To(Say("Application lifecycle:"))
			Expect(testUI.Out).To(Say("  apps,a\\s+run-task,rt\\s+events"))
			Expect(testUI.Out).To(Say("  restage,rg\\s+scale"))

			Expect(testUI.Out).To(Say("Services integration:"))
			Expect(testUI.Out).To(Say("  marketplace,m\\s+create-user-provided-service,cups"))
			Expect(testUI.Out).To(Say("  services,s\\s+update-user-provided-service,uups"))

			Expect(testUI.Out).To(Say("Route and domain management:"))
			Expect(testUI.Out).To(Say("  routes,r\\s+delete-route\\s+create-domain"))
			Expect(testUI.Out).To(Say("  domains\\s+map-route"))

			Expect(testUI.Out).To(Say("Space management:"))
			Expect(testUI.Out).To(Say("  spaces\\s+create-space\\s+set-space-role"))

			Expect(testUI.Out).To(Say("Org management:"))
			Expect(testUI.Out).To(Say("  orgs,o\\s+set-org-role"))

			Expect(testUI.Out).To(Say("CLI plugin management:"))
			Expect(testUI.Out).To(Say("  install-plugin    list-plugin-repos"))

			Expect(testUI.Out).To(Say("Global options:"))
			Expect(testUI.Out).To(Say("  --help, -h                         Show help"))
			Expect(testUI.Out).To(Say("  -v                                 Print API request diagnostics to stdout"))

			Expect(testUI.Out).To(Say("Use 'cf help -a' to see all commands\\."))
		})

		Context("when there are multiple installed plugins", func() {
			BeforeEach(func() {
				fakeConfig.PluginsReturns([]configv3.Plugin{
					{
						Name: "Some-other-plugin",
						Commands: []configv3.PluginCommand{
							{
								Name:     "some-other-plugin-command",
								HelpText: "does some other thing",
							},
						},
					},
					{
						Name: "some-plugin",
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
					{
						Name: "the-last-plugin",
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

				Expect(testUI.Out).To(Say("Commands offered by installed plugins:"))
				Expect(testUI.Out).To(Say("some-other-plugin-command\\s+enable\\s+last-plugin-command"))
				Expect(testUI.Out).To(Say("disable\\s+some-other-command"))

			})
		})
	})

	Describe("providing help for all commands", func() {
		Context("when a command is not provided", func() {
			BeforeEach(func() {
				cmd.OptionalArgs = flag.CommandName{
					CommandName: "",
				}
				cmd.AllCommands = true

				cmd.Actor = sharedaction.NewActor(nil)
				fakeConfig.PluginsReturns([]configv3.Plugin{
					{
						Name: "Diego-Enabler",
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

				Expect(testUI.Out).To(Say("NAME:"))
				Expect(testUI.Out).To(Say("   faceman - A command line tool to interact with Cloud Foundry"))
				Expect(testUI.Out).To(Say("USAGE:"))
				Expect(testUI.Out).To(Say("   faceman \\[global options\\] command \\[arguments...\\] \\[command options\\]"))
				Expect(testUI.Out).To(Say("VERSION:"))
				Expect(testUI.Out).To(Say("   face2.0-yesterday"))

				Expect(testUI.Out).To(Say("GETTING STARTED:"))
				Expect(testUI.Out).To(Say("   help\\s+Show help"))
				Expect(testUI.Out).To(Say("   api\\s+Set or view target api url"))

				Expect(testUI.Out).To(Say("APPS:"))
				Expect(testUI.Out).To(Say("   apps\\s+List all apps in the target space"))
				Expect(testUI.Out).To(Say("   restart-app-instance\\s+Terminate, then restart an app instance"))
				Expect(testUI.Out).To(Say("   ssh-enabled\\s+Reports whether SSH is enabled on an application container instance"))

				Expect(testUI.Out).To(Say("SERVICES:"))
				Expect(testUI.Out).To(Say("   marketplace\\s+List available offerings in the marketplace"))
				Expect(testUI.Out).To(Say("   create-service\\s+Create a service instance"))

				Expect(testUI.Out).To(Say("ORGS:"))
				Expect(testUI.Out).To(Say("   orgs\\s+List all orgs"))
				Expect(testUI.Out).To(Say("   delete-org\\s+Delete an org"))

				Expect(testUI.Out).To(Say("SPACES:"))
				Expect(testUI.Out).To(Say("   spaces\\s+List all spaces in an org"))
				Expect(testUI.Out).To(Say("   allow-space-ssh\\s+Allow SSH access for the space"))

				Expect(testUI.Out).To(Say("DOMAINS:"))
				Expect(testUI.Out).To(Say("   domains\\s+List domains in the target org"))
				Expect(testUI.Out).To(Say("   router-groups\\s+List router groups"))

				Expect(testUI.Out).To(Say("ROUTES:"))
				Expect(testUI.Out).To(Say("   routes\\s+List all routes in the current space or the current organization"))
				Expect(testUI.Out).To(Say("   unmap-route\\s+Remove a url route from an app"))

				Expect(testUI.Out).To(Say("NETWORK POLICIES:"))
				Expect(testUI.Out).To(Say("   network-policies\\s+List direct network traffic policies"))
				Expect(testUI.Out).To(Say("   add-network-policy\\s+Create policy to allow direct network traffic from one app to another"))
				Expect(testUI.Out).To(Say("   remove-network-policy\\s+Remove network traffic policy of an app"))

				Expect(testUI.Out).To(Say("BUILDPACKS:"))
				Expect(testUI.Out).To(Say("   buildpacks\\s+List all buildpacks"))
				Expect(testUI.Out).To(Say("   delete-buildpack\\s+Delete a buildpack"))

				Expect(testUI.Out).To(Say("USER ADMIN:"))
				Expect(testUI.Out).To(Say("   create-user\\s+Create a new user"))
				Expect(testUI.Out).To(Say("   space-users\\s+Show space users by role"))

				Expect(testUI.Out).To(Say("ORG ADMIN:"))
				Expect(testUI.Out).To(Say("   quotas\\s+List available usage quotas"))
				Expect(testUI.Out).To(Say("   delete-quota\\s+Delete a quota"))

				Expect(testUI.Out).To(Say("SPACE ADMIN:"))
				Expect(testUI.Out).To(Say("   space-quotas\\s+List available space resource quotas"))
				Expect(testUI.Out).To(Say("   set-space-quota\\s+Assign a space quota definition to a space"))

				Expect(testUI.Out).To(Say("SERVICE ADMIN:"))
				Expect(testUI.Out).To(Say("   service-auth-tokens\\s+List service auth tokens"))
				Expect(testUI.Out).To(Say("   service-access\\s+List service access settings"))

				Expect(testUI.Out).To(Say("SECURITY GROUP:"))
				Expect(testUI.Out).To(Say("   security-group\\s+Show a single security group"))
				Expect(testUI.Out).To(Say("   staging-security-groups\\s+List security groups in the staging set for applications"))

				Expect(testUI.Out).To(Say("ENVIRONMENT VARIABLE GROUPS:"))
				Expect(testUI.Out).To(Say("   running-environment-variable-group\\s+Retrieve the contents of the running environment variable group"))
				Expect(testUI.Out).To(Say("   set-running-environment-variable-group Pass parameters as JSON to create a running environment variable group"))

				Expect(testUI.Out).To(Say("ISOLATION SEGMENTS:"))
				Expect(testUI.Out).To(Say("   isolation-segments\\s+List all isolation segments"))
				Expect(testUI.Out).To(Say("   create-isolation-segment\\s+Create an isolation segment"))
				Expect(testUI.Out).To(Say("   delete-isolation-segment\\s+Delete an isolation segment"))
				Expect(testUI.Out).To(Say("   enable-org-isolation\\s+Entitle an organization to an isolation segment"))
				Expect(testUI.Out).To(Say("   disable-org-isolation\\s+Revoke an organization's entitlement to an isolation segment"))
				Expect(testUI.Out).To(Say("   set-org-default-isolation-segment\\s+Set the default isolation segment used for apps in spaces in an org"))
				Expect(testUI.Out).To(Say("   reset-org-default-isolation-segment\\s+Reset the default isolation segment used for apps in spaces of an org"))
				Expect(testUI.Out).To(Say("   set-space-isolation-segment"))
				Expect(testUI.Out).To(Say("   reset-space-isolation-segment"))

				Expect(testUI.Out).To(Say("FEATURE FLAGS:"))
				Expect(testUI.Out).To(Say("   feature-flags\\s+Retrieve list of feature flags with status"))
				Expect(testUI.Out).To(Say("   disable-feature-flag"))

				Expect(testUI.Out).To(Say("ADVANCED:"))
				Expect(testUI.Out).To(Say("   curl\\s+Executes a request to the targeted API endpoint"))
				Expect(testUI.Out).To(Say("   ssh-code\\s+Get a one time password for ssh clients"))

				Expect(testUI.Out).To(Say("ADD/REMOVE PLUGIN REPOSITORY:"))
				Expect(testUI.Out).To(Say("   add-plugin-repo\\s+Add a new plugin repository"))
				Expect(testUI.Out).To(Say("   repo-plugins\\s+List all available plugins in specified repository or in all added repositories"))

				Expect(testUI.Out).To(Say("ADD/REMOVE PLUGIN:"))
				Expect(testUI.Out).To(Say("   plugins\\s+List commands of installed plugins"))
				Expect(testUI.Out).To(Say("   uninstall-plugin\\s+Uninstall CLI plugin"))

				Expect(testUI.Out).To(Say("INSTALLED PLUGIN COMMANDS:"))
				Expect(testUI.Out).To(Say("   enable-diego\\s+enable Diego support for an app"))

				Expect(testUI.Out).To(Say("ENVIRONMENT VARIABLES:"))
				Expect(testUI.Out).To(Say("   CF_COLOR=false                     Do not colorize output"))
				Expect(testUI.Out).To(Say("   CF_DIAL_TIMEOUT=5                  Max wait time to establish a connection, including name resolution, in seconds"))
				Expect(testUI.Out).To(Say("   CF_HOME=path/to/dir/               Override path to default config directory"))
				Expect(testUI.Out).To(Say("   CF_PLUGIN_HOME=path/to/dir/        Override path to default plugin config directory"))
				Expect(testUI.Out).To(Say("   CF_TRACE=true                      Print API request diagnostics to stdout"))
				Expect(testUI.Out).To(Say("   CF_TRACE=path/to/trace.log         Append API request diagnostics to a log file"))
				Expect(testUI.Out).To(Say("   https_proxy=proxy.example.com:8080 Enable HTTP proxying for API requests"))

				Expect(testUI.Out).To(Say("GLOBAL OPTIONS:"))
				Expect(testUI.Out).To(Say("   --help, -h                         Show help"))
				Expect(testUI.Out).To(Say("   -v                                 Print API request diagnostics to stdout"))

				Expect(testUI.Out).To(Say("APPS \\(experimental\\):"))
				Expect(testUI.Out).To(Say("   v3-apps\\s+List all apps in the target space"))
				Expect(testUI.Out).To(Say("   v3-app\\s+Display health and status for an app"))
				Expect(testUI.Out).To(Say("   v3-create-app\\s+Create a V3 App"))
				Expect(testUI.Out).To(Say("   v3-push\\s+Push a new app or sync changes to an existing app"))
				Expect(testUI.Out).To(Say("   v3-scale\\s+Change or view the instance count, disk space limit, and memory limit for an app"))
				Expect(testUI.Out).To(Say("   v3-delete\\s+Delete a V3 App"))
				Expect(testUI.Out).To(Say("   v3-start\\s+Start an app"))
				Expect(testUI.Out).To(Say("   v3-stop\\s+Stop an app"))
				Expect(testUI.Out).To(Say("   v3-restart\\s+Stop all instances of the app, then start them again. This causes downtime."))
				Expect(testUI.Out).To(Say("   v3-stage\\s+Create a new droplet for an app"))
				Expect(testUI.Out).To(Say("   v3-restart-app-instance\\s+Terminate, then instantiate an app instance"))
				Expect(testUI.Out).To(Say("   v3-droplets\\s+List droplets of an app"))
				Expect(testUI.Out).To(Say("   v3-set-droplet\\s+Set the droplet used to run an app"))
				Expect(testUI.Out).To(Say("   v3-set-env\\s+Set an env variable for an app"))
				Expect(testUI.Out).To(Say("   v3-unset-env\\s+Remove an env variable from an app"))
				Expect(testUI.Out).To(Say("   v3-get-health-check\\s+Show the type of health check performed on an app"))
				Expect(testUI.Out).To(Say("   v3-set-health-check\\s+Change type of health check performed on an app's process"))
				Expect(testUI.Out).To(Say("   v3-packages\\s+List packages of an app"))
				Expect(testUI.Out).To(Say("   v3-create-package\\s+Uploads a V3 Package"))
			})

			Context("when there are multiple installed plugins", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{
						{
							Name: "Some-other-plugin",
							Commands: []configv3.PluginCommand{
								{
									Name:     "some-other-plugin-command",
									HelpText: "does some other thing",
								},
							},
						},
						{
							Name: "some-plugin",
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
						{
							Name: "the-last-plugin",
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

					Expect(testUI.Out).To(Say(`INSTALLED PLUGIN COMMANDS:.*
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
