package options_test

import (
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHOptions", func() {
	var (
		opts       *options.SSHOptions
		args       []string
		parseError error
	)

	Describe("Parse", func() {
		BeforeEach(func() {
			opts = options.NewSSHOptions()
			args = []string{"ssh"}
			parseError = nil
		})

		JustBeforeEach(func() {
			parseError = opts.Parse(args)
		})

		Context("when the command name is missing", func() {
			BeforeEach(func() {
				args = []string{}
			})

			It("returns a UsageError", func() {
				Expect(parseError).To(Equal(options.UsageError))
			})
		})

		Context("when the wrong command name is present", func() {
			BeforeEach(func() {
				args = []string{"scp"}
			})

			It("returns a UsageError", func() {
				Expect(parseError).To(Equal(options.UsageError))
			})
		})

		Context("when no arguments are specified", func() {
			It("returns a UsageError", func() {
				Expect(parseError).To(Equal(options.UsageError))
			})
		})

		Context("when no app name is provided", func() {
			BeforeEach(func() {
				args = append(args, "-i", "3")
			})

			It("returns a UsageError", func() {
				Expect(parseError).To(Equal(options.UsageError))
			})
		})

		Context("when an app name is provided", func() {
			Context("as the only argument", func() {
				BeforeEach(func() {
					args = append(args, "app-1")
				})

				It("populates the AppName field", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.AppName).To(Equal("app-1"))
				})
			})

			Context("as the last argument", func() {
				BeforeEach(func() {
					args = append(args, "-i", "3", "app-1")
				})

				It("populates the AppName field", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.AppName).To(Equal("app-1"))
				})
			})
		})

		Context("when --skip-host-validation is set", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "--skip-host-validation")
			})

			It("disables host key validation", func() {
				Expect(parseError).ToNot(HaveOccurred())
				Expect(opts.SkipHostValidation).To(BeTrue())
				Expect(opts.AppName).To(Equal("app-name"))
			})
		})

		Context("when -k is set", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-k")
			})

			It("disables host key validation", func() {
				Expect(parseError).ToNot(HaveOccurred())
				Expect(opts.SkipHostValidation).To(BeTrue())
				Expect(opts.AppName).To(Equal("app-name"))
			})
		})

		Context("when an -i flag is provided", func() {
			BeforeEach(func() {
				args = append(args, "app-name")
			})

			Context("without an argument", func() {
				BeforeEach(func() {
					args = append(args, "-i")
				})

				It("returns an error", func() {
					Expect(parseError).To(MatchError("missing parameter for -i"))
				})
			})

			Context("with a positive integer argument", func() {
				BeforeEach(func() {
					args = append(args, "-i", "3")
				})

				It("populates the Index field", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.Index).To(BeEquivalentTo(3))
				})
			})

			Context("with a negative integer argument", func() {
				BeforeEach(func() {
					args = append(args, "-i", "-3")
				})

				It("returns an error", func() {
					Expect(parseError).To(MatchError("not a valid number: -3"))
				})
			})

			Context("with a non-numeric argument", func() {
				BeforeEach(func() {
					args = append(args, "-i", "three")
				})

				It("returns an error", func() {
					Expect(parseError).To(MatchError("not a valid number: three"))
				})
			})
		})

		Context("when the -t and -T flags are not used", func() {
			BeforeEach(func() {
				args = append(args, "app-name")
			})

			It("requests auto tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_AUTO))
			})
		})

		Context("when the -T flag is provided", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-T")
			})

			It("disables tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_NO))
			})
		})

		Context("when the -t flag is used once", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-t")
			})

			It("requests tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_YES))
			})
		})

		Context("when the -t flag is used more than once", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-tt")
			})
			It("foces tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_FORCE))
			})
		})

		Context("when -t and -T are both specified", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-tTt")
			})

			It("disables tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_NO))
			})
		})

		Context("when a command is specified", func() {
			Context("without flags", func() {
				BeforeEach(func() {
					args = append(args, "app-name", "-k", "-t", "true")
				})

				It("handles the app and command correctly", func() {
					Expect(opts.SkipHostValidation).To(BeTrue())
					Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_YES))
					Expect(opts.AppName).To(Equal("app-name"))
					Expect(opts.Command).To(ConsistOf("true"))
				})
			})

			Context("with flags", func() {
				BeforeEach(func() {
					args = append(args, "-k", "app-name", "-t", "echo", "-n", "hello!")
				})

				It("handles the app and command correctly", func() {
					Expect(opts.SkipHostValidation).To(BeTrue())
					Expect(opts.TerminalRequest).To(Equal(options.REQUEST_TTY_YES))
					Expect(opts.AppName).To(Equal("app-name"))
					Expect(opts.Command).To(ConsistOf("echo", "-n", "hello!"))
				})
			})
		})

		Context("when local port forwarding is requested", func() {
			BeforeEach(func() {
				args = append(args, "app-name")
			})

			Context("without an explicit bind address", func() {
				BeforeEach(func() {
					args = append(args, "-L", "9999:remote:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: "localhost:9999", ConnectAddress: "remote:8888"}))
				})
			})

			Context("with an explit bind address", func() {
				BeforeEach(func() {
					args = append(args, "-L", "explicit:9999:remote:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: "explicit:9999", ConnectAddress: "remote:8888"}))
				})
			})

			Context("with an explicit ipv6 bind address", func() {
				BeforeEach(func() {
					args = append(args, "-L", "[::]:9999:remote:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: "[::]:9999", ConnectAddress: "remote:8888"}))
				})
			})

			Context("with an empty bind address", func() {
				BeforeEach(func() {
					args = append(args, "-L", ":9999:remote:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: ":9999", ConnectAddress: "remote:8888"}))
				})
			})

			Context("with * as the bind address", func() {
				BeforeEach(func() {
					args = append(args, "-L", "*:9999:remote:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: ":9999", ConnectAddress: "remote:8888"}))
				})
			})

			Context("with an explicit ipv6 connect address", func() {
				BeforeEach(func() {
					args = append(args, "-L", "[::]:9999:[2001:db8::1]:8888")
				})

				It("sets the forward spec", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(options.ForwardSpec{ListenAddress: "[::]:9999", ConnectAddress: "[2001:db8::1]:8888"}))
				})
			})

			Context("with a missing bracket", func() {
				BeforeEach(func() {
					args = append(args, "-L", "localhost:9999:[example.com:8888")
				})

				It("returns an error", func() {
					Expect(parseError).To(MatchError(`Argument missing closing bracket: "[example.com:8888"`))
				})
			})

			Context("when a closing bracket is not followed by a colon", func() {
				BeforeEach(func() {
					args = append(args, "-L", "localhost:9999:[example.com]8888")
				})

				It("returns an error", func() {
					Expect(parseError).To(MatchError(`Unexpected token: "8888"`))
				})
			})

			Context("when multiple local port forward options are specified", func() {
				BeforeEach(func() {
					args = append(args, "-L", "9999:remote:8888")
					args = append(args, "-L8080:remote:80")
				})

				It("sets the forward specs", func() {
					Expect(parseError).NotTo(HaveOccurred())
					Expect(opts.ForwardSpecs).To(ConsistOf(
						options.ForwardSpec{ListenAddress: "localhost:9999", ConnectAddress: "remote:8888"},
						options.ForwardSpec{ListenAddress: "localhost:8080", ConnectAddress: "remote:80"},
					))
				})
			})
		})

		Context("when -N is specified", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-N")
			})

			It("indicates that no remote command should be run", func() {
				Expect(parseError).ToNot(HaveOccurred())
				Expect(opts.SkipRemoteExecution).To(BeTrue())
				Expect(opts.AppName).To(Equal("app-name"))
			})
		})
	})

	Describe("SSHUsage", func() {
		It("prints usage information", func() {
			usage := options.SSHUsage()

			Expect(usage).To(ContainSubstring("Usage: ssh [-kNTt] [-i app-instance-index] [-L [bind_address:]port:host:hostport] app-name [command]"))
			Expect(usage).To(ContainSubstring("-i, --index=app-instance-index"))
			Expect(usage).To(ContainSubstring("-k, --skip-host-validation"))
			Expect(usage).To(ContainSubstring("-L [bind_address:]port:host:hostport"))
			Expect(usage).To(ContainSubstring("-N    do not execute a remote command"))
			Expect(usage).To(ContainSubstring("-T    disable pseudo-tty allocation"))
			Expect(usage).To(ContainSubstring("-t    force pseudo-tty allocation"))
		})
	})
})
