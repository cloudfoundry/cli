package experimental

import (
	"fmt"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-ssh command", func() {
	var (
		appName   string
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Context("when --help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("v3-ssh", "--help")

			Eventually(session.Out).Should(Say(`NAME:`))
			Eventually(session.Out).Should(Say(`ssh - SSH to an application container instance`))
			Eventually(session.Out).Should(Say(`USAGE:`))
			Eventually(session.Out).Should(Say(`cf v3-ssh APP_NAME \[--process PROCESS\] \[-i INDEX\] \[-c COMMAND\]\n`))
			Eventually(session.Out).Should(Say(`\[-L \[BIND_ADDRESS:\]LOCAL_PORT:REMOTE_HOST:REMOTE_PORT\]\.\.\. \[--skip-remote-execution\]`))
			Eventually(session.Out).Should(Say(`\[--disable-pseudo-tty \| --force-pseudo-tty \| --request-pseudo-tty\] \[--skip-host-validation\]`))
			Eventually(session.Out).Should(Say(`OPTIONS:`))
			Eventually(session.Out).Should(Say(`--app-instance-index, -i\s+App process instance index \(Default: 0\)`))
			Eventually(session.Out).Should(Say(`--command, -c\s+Command to run`))
			Eventually(session.Out).Should(Say(`--disable-pseudo-tty, -T\s+Disable pseudo-tty allocation`))
			Eventually(session.Out).Should(Say(`--force-pseudo-tty\s+Force pseudo-tty allocation`))
			Eventually(session.Out).Should(Say(`-L\s+Local port forward specification`))
			Eventually(session.Out).Should(Say(`--process\s+App process name \(Default: web\)`))
			Eventually(session.Out).Should(Say(`--request-pseudo-tty, -t\s+Request pseudo-tty allocation`))
			Eventually(session.Out).Should(Say(`--skip-host-validation, -k\s+Skip host key validation\. Not recommended!`))
			Eventually(session.Out).Should(Say(`--skip-remote-execution, -N\s+Do not execute a remote command`))
			Eventually(session.Out).Should(Say(`SEE ALSO:`))
			Eventually(session.Out).Should(Say(`allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled`))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-ssh")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-ssh", appName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		BeforeEach(func() {
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app does not exist", func() {
			It("it displays the app does not exist", func() {
				session := helpers.CF("v3-ssh", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			Context("TTY Options", func() {
				// * The columns specify the various TTY flags passed to cf ssh
				//   (--disable-pseudo-tty, --force-pseudo-tty, --request-pseudo-tty).
				// * The rows specify what kind of shell you’re running "cf ssh" from. To
				//   simulate an interactive shell, simply use your terminal as always.
				//   To simulate a non-interactive shell, append "<< EOF <new-line>
				//   <command-to-execute-on-remote-host> <new-line> EOF" to your command
				// * The values (yes/no) determine whether a TTY session should be
				//   allocated on the remote host. Verify by running "TTY" on remote host.
				//
				//               TTY Option -> | Default(auto) | Disable | Force | Request
				// Shell_Type__________________|_______________|_________|_______|_____________
				// interactive                 | Yes           | No      | Yes   | Yes
				// non-interactive             | No            | No      | No    | No
				// interactive w/ commands     | No            | No      | Yes   | Yes
				// non-interactive w/ commands | No            | No      | Yes   | No

				Context("when the running session is interactive", func() {
					// This should be tested manually (launching an interactive shell in code is hard)
				})

				Context("when the running session is non-interactive", func() {
					Context("when providing commands to run on the remote host", func() {
						Context("when using default tty option (auto)", func() {
							It("the remote shell is not TTY", func() {
								// we echo hello because a successful ssh call returns the status
								session := helpers.CF("v3-ssh", appName, "-c tty;", "-c echo hello")
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when disable-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								session := helpers.CF("v3-ssh", appName, "--disable-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when force-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								session := helpers.CF("v3-ssh", appName, "--force-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session.Out).ShouldNot(Say("not a tty"))
								Eventually(session.Out).Should(Say("/dev/*"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when request-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								session := helpers.CF("v3-ssh", appName, "--request-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when not providing commands as args", func() {
						var buffer *Buffer

						BeforeEach(func() {
							buffer = NewBuffer()
						})

						Context("when using default tty option (auto)", func() {
							It("the remote shell is not TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "v3-ssh", appName)
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when disable-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "v3-ssh", appName, "--disable-pseudo-tty")
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when force-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "v3-ssh", appName, "--force-pseudo-tty")
								Eventually(session.Out).ShouldNot(Say("not a tty"))
								Eventually(session.Out).Should(Say("/dev/*"))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when request-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "v3-ssh", appName, "--request-pseudo-tty")
								Eventually(session.Out).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})

			It("ssh's to the process 'web', index '0'", func() {
				session := helpers.CF("v3-ssh", appName, "-c", "ps aux;", "-c", "env")
				// To verify we ssh'd into the web process we examine processes
				// that were launched tha are unique to that process
				Eventually(session.Out).Should(Say("vcap.*ruby"))
				Eventually(session.Out).Should(Say("INSTANCE_INDEX=0"))
				Eventually(session).Should(Exit(0))
			})

			Context("when commands to run are specified", func() {
				It("ssh's to the default container and runs the commands", func() {
					session := helpers.CF("v3-ssh", appName, "-c", "ls;", "-c", "echo $USER")
					Eventually(session.Out).Should(Say("app"))
					Eventually(session.Out).Should(Say("deps"))
					Eventually(session.Out).Should(Say("logs"))
					Eventually(session.Out).Should(Say("vcap"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the application hasn't started", func() {
				BeforeEach(func() {
					session := helpers.CF("v3-stop", appName)
					Eventually(session).Should(Exit(0))
				})

				It("prints an error message", func() {
					session := helpers.CF("v3-ssh", appName)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(fmt.Sprintf("Application '%s' is not in the STARTED state", appName)))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the remote command exits with a different status code", func() {
				It("exits with that status code", func() {
					session := helpers.CF("v3-ssh", appName, "-c", "asdf")
					Eventually(session).Should(Exit(127))
				})
			})

			Context("when port forwarding is used", func() {
				var port int

				BeforeEach(func() {
					port = 55500 + GinkgoParallelNode()
				})

				It("configures local port to connect to the app port", func() {
					session := helpers.CF("v3-ssh", appName, "-N", "-L", fmt.Sprintf("%d:localhost:8080", port))

					time.Sleep(5 * time.Second) // Need to wait a few seconds for pipes to connect.
					curl, err := Start(
						exec.Command("curl", fmt.Sprintf("localhost:%d/", port)),
						GinkgoWriter,
						GinkgoWriter,
					)
					Expect(err).ToNot(HaveOccurred())

					Eventually(curl).Should(Say("WEBrick"))
					Eventually(curl).Should(Exit(0))

					session.Kill()
					Eventually(session).Should(Exit())
				})
			})

			Context("when a process is specified", func() {
				Context("when the process does not exist", func() {
					It("displays the process does not exist", func() {
						session := helpers.CF("v3-ssh", appName, "--process", "fake-process")
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Process fake-process not found"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the process exists", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("v3-scale", appName, "--process", "console", "-i", "1")).Should(Exit(0))
					})

					It("ssh's to the process's default index", func() {
						session := helpers.CF("v3-ssh", appName, "--process", "console", "-c", "ps aux;", "-c", "env")
						Eventually(session.Out).Should(Say("vcap.*irb"))
						Eventually(session.Out).Should(Say("INSTANCE_INDEX=0"))
						Eventually(session).Should(Exit(0))
					})

					Context("when the index is specified", func() {
						Context("when the index does not exist", func() {
							It("returns an instance not found error", func() {
								session := helpers.CF("v3-ssh", appName, "--process", "console", "-i", "1", "-c", "ps aux;", "-c", "env")
								Eventually(session.Out).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Instance %d of process console not found", 1))
								Eventually(session).Should(Exit(1))
							})
						})

						Context("when the index exists", func() {
							It("ssh's to the provided index", func() {
								session := helpers.CF("v3-ssh", appName, "--process", "console", "-i", "0", "-c", "ps aux;", "-c", "env")
								Eventually(session.Out).Should(Say("vcap.*irb"))
								Eventually(session.Out).Should(Say("INSTANCE_INDEX=0"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})

			Context("when a user isn't authorized", func() {
				var (
					newUser string
					newPass string
				)

				BeforeEach(func() {
					newUser = helpers.NewUsername()
					newPass = helpers.NewPassword()

					Eventually(helpers.CF("create-user", newUser, newPass)).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", newUser, orgName, spaceName, "SpaceAuditor")).Should(Exit(0))
					Eventually(helpers.CF("auth", newUser, newPass)).Should(Exit(0))
					helpers.TargetOrgAndSpace(orgName, spaceName)
				})

				AfterEach(func() {
					helpers.LoginCF()
				})

				It("returns an error", func() {
					session := helpers.CF("v3-ssh", appName)

					Eventually(session.Err).Should(Say("Error opening SSH connection: You are not authorized to perform the requested action."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
