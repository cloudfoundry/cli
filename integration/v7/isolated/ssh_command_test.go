package isolated

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh command", func() {
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

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("ssh", "--help")

			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`ssh - SSH to an application container instance`))
			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf ssh APP_NAME \[--process PROCESS\] \[-i INDEX\] \[-c COMMAND\]...\n`))
			Eventually(session).Should(Say(`\[-L \[BIND_ADDRESS:\]LOCAL_PORT:REMOTE_HOST:REMOTE_PORT\]\.\.\. \[--skip-remote-execution\]`))
			Eventually(session).Should(Say(`\[--disable-pseudo-tty \| --force-pseudo-tty \| --request-pseudo-tty\] \[--skip-host-validation\]`))
			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--app-instance-index, -i\s+App process instance index \(Default: 0\)`))
			Eventually(session).Should(Say(`--command, -c\s+Command to run`))
			Eventually(session).Should(Say(`--disable-pseudo-tty, -T\s+Disable pseudo-tty allocation`))
			Eventually(session).Should(Say(`--force-pseudo-tty\s+Force pseudo-tty allocation`))
			Eventually(session).Should(Say(`-L\s+Local port forward specification`))
			Eventually(session).Should(Say(`--process\s+App process name \(Default: web\)`))
			Eventually(session).Should(Say(`--request-pseudo-tty, -t\s+Request pseudo-tty allocation`))
			Eventually(session).Should(Say(`--skip-host-validation, -k\s+Skip host key validation\. Not recommended!`))
			Eventually(session).Should(Say(`--skip-remote-execution, -N\s+Do not execute a remote command`))
			Eventually(session).Should(Say(`ENVIRONMENT:`))
			Eventually(session).Should(Say(`all_proxy=\s+Specify a proxy server to enable proxying for all requests`))
			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("ssh")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "ssh", appName)
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("it displays the app does not exist", func() {
				session := helpers.CF("ssh", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
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

				When("the running session is interactive", func() {
					// This should be tested manually (launching an interactive shell in code is hard)
				})

				When("the running session is non-interactive", func() {
					When("providing commands to run on the remote host", func() {
						When("using default tty option (auto)", func() {
							It("the remote shell is not TTY", func() {
								// we echo hello because a successful ssh call returns the status
								session := helpers.CF("ssh", appName, "-c tty;", "-c echo hello")
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("disable-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								session := helpers.CF("ssh", appName, "--disable-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("force-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								session := helpers.CF("ssh", appName, "--force-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session).ShouldNot(Say("not a tty"))
								Eventually(session).Should(Say("/dev/*"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("request-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								session := helpers.CF("ssh", appName, "--request-pseudo-tty", "-c tty;", "-c echo hello")
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("not providing commands as args", func() {
						var buffer *Buffer

						BeforeEach(func() {
							buffer = NewBuffer()
						})

						When("using default tty option (auto)", func() {
							It("the remote shell is not TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "ssh", appName)
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("disable-pseudo-tty is specified", func() {
							It("the remote shell is not TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "ssh", appName, "--disable-pseudo-tty")
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("force-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "ssh", appName, "--force-pseudo-tty")
								Eventually(session).ShouldNot(Say("not a tty"))
								Eventually(session).Should(Say("/dev/*"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("request-pseudo-tty is specified", func() {
							It("the remote shell is TTY", func() {
								buffer.Write([]byte("tty\n"))
								buffer.Write([]byte("echo hello\n"))
								buffer.Write([]byte("exit\n"))
								session := helpers.CFWithStdin(buffer, "ssh", appName, "--request-pseudo-tty")
								Eventually(session).Should(Say("not a tty"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})

			It("ssh's to the process 'web', index '0'", func() {
				session := helpers.CF("ssh", appName, "-c", "ps aux;", "-c", "env")
				// To verify we ssh'd into the web process we examine processes
				// that were launched that are unique to that process
				Eventually(session).Should(Say("vcap.*ruby"))
				Eventually(session).Should(Say("INSTANCE_INDEX=0"))
				Eventually(session).Should(Exit(0))
			})

			When("commands to run are specified", func() {
				It("ssh's to the default container and runs the commands", func() {
					session := helpers.CF("ssh", appName, "-c", "ls;", "-c", "echo $USER")
					Eventually(session).Should(Say("app"))
					Eventually(session).Should(Say("deps"))
					Eventually(session).Should(Say("logs"))
					Eventually(session).Should(Say("vcap"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the application hasn't started", func() {
				BeforeEach(func() {
					session := helpers.CF("v3-stop", appName)
					Eventually(session).Should(Exit(0))
				})

				It("prints an error message", func() {
					session := helpers.CF("ssh", appName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(fmt.Sprintf("Application '%s' is not in the STARTED state", appName)))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the remote command exits with a different status code", func() {
				It("exits with that status code", func() {
					session := helpers.CF("ssh", appName, "-c", "asdf")
					Eventually(session).Should(Exit(127))
				})
			})

			When("port forwarding is used", func() {
				var port int

				BeforeEach(func() {
					port = 55500 + GinkgoParallelNode()
				})

				It("configures local port to connect to the app port", func() {
					session := helpers.CF("ssh", appName, "-N", "-L", fmt.Sprintf("%d:localhost:8080", port))

					time.Sleep(35 * time.Second) // Need to wait a few seconds for pipes to connect.
					response, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
					Expect(err).ToNot(HaveOccurred())
					defer response.Body.Close()

					Eventually(BufferReader(response.Body)).Should(Say("WEBrick"))

					session.Kill()
					Eventually(session).Should(Exit())
				})
			})

			When("a process is specified", func() {
				When("the process does not exist", func() {
					It("displays the process does not exist", func() {
						session := helpers.CF("ssh", appName, "--process", "fake-process")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Process fake-process not found"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the process exists", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("scale", appName, "--process", "console", "-i", "1")).Should(Exit(0))
					})

					It("ssh's to the process's default index", func() {
						session := helpers.CF("ssh", appName, "--process", "console", "-c", "ps aux;", "-c", "env")
						Eventually(session).Should(Say("vcap.*irb"))
						Eventually(session).Should(Say("INSTANCE_INDEX=0"))
						Eventually(session).Should(Exit(0))
					})

					When("the index is specified", func() {
						When("the index does not exist", func() {
							It("returns an instance not found error", func() {
								session := helpers.CF("ssh", appName, "--process", "console", "-i", "1", "-c", "ps aux;", "-c", "env")
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Instance %d of process console not found", 1))
								Eventually(session).Should(Exit(1))
							})
						})

						When("the index exists", func() {
							It("ssh's to the provided index", func() {
								session := helpers.CF("ssh", appName, "--process", "console", "-i", "0", "-c", "ps aux;", "-c", "env")
								Eventually(session).Should(Say("vcap.*irb"))
								Eventually(session).Should(Say("INSTANCE_INDEX=0"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})

			When("a user isn't authorized", func() {
				var (
					newUser string
					newPass string
				)

				BeforeEach(func() {
					newUser = helpers.NewUsername()
					newPass = helpers.NewPassword()

					Eventually(helpers.CF("create-user", newUser, newPass)).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", newUser, orgName, spaceName, "SpaceAuditor")).Should(Exit(0))
					env := map[string]string{
						"CF_USERNAME": newUser,
						"CF_PASSWORD": newPass,
					}
					Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
					helpers.TargetOrgAndSpace(orgName, spaceName)
				})

				AfterEach(func() {
					helpers.LoginCF()
				})

				It("returns an error", func() {
					session := helpers.CF("ssh", appName)

					Eventually(session.Err).Should(Say("Error opening SSH connection: You are not authorized to perform the requested action."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
