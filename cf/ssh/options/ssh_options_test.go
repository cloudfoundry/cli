package options_test

import (
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/ssh/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHOptions", func() {
	var (
		opts       *options.SSHOptions
		args       []string
		parseError error
		fc         flags.FlagContext
	)

	Describe("Parse", func() {
		BeforeEach(func() {
			fc = flags.New()
			fc.NewStringSliceFlag("L", "", "")
			fc.NewStringSliceFlag("command", "c", "")
			fc.NewIntFlag("app-instance-index", "i", "")
			fc.NewBoolFlag("skip-host-validation", "k", "")
			fc.NewBoolFlag("skip-remote-execution", "N", "")
			fc.NewBoolFlag("request-pseudo-tty", "t", "")
			fc.NewBoolFlag("force-pseudo-tty", "tt", "")
			fc.NewBoolFlag("disable-pseudo-tty", "T", "")

			args = []string{}
			parseError = nil
		})

		JustBeforeEach(func() {
			err := fc.Parse(args...)
			Expect(err).NotTo(HaveOccurred())

			opts, parseError = options.NewSSHOptions(fc)
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

		Context("when the -t and -T flags are not used", func() {
			BeforeEach(func() {
				args = append(args, "app-name")
			})

			It("requests auto tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYAuto))
			})
		})

		Context("when the -T flag is provided", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-T")
			})

			It("disables tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYNo))
			})
		})

		Context("when the -t flag is used", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-t")
			})

			It("requests tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYYes))
			})
		})

		Context("when the -tt flag is used", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-tt")
			})
			It("foces tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYForce))
			})
		})

		Context("when both -t, -tt are specified", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-t", "-tt")
			})

			It("forces tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYForce))
			})
		})

		Context("when -t, -tt and -T are all specified", func() {
			BeforeEach(func() {
				args = append(args, "app-name", "-t", "-tt", "-T")
			})

			It("disables tty allocation", func() {
				Expect(opts.TerminalRequest).To(Equal(options.RequestTTYNo))
			})
		})

		Context("when command is provided with -c", func() {
			Context("when -c is used once", func() {
				BeforeEach(func() {
					args = append(args, "app-name", "-k", "-t", "-c", "true")
				})

				It("handles the app and command correctly", func() {
					Expect(opts.SkipHostValidation).To(BeTrue())
					Expect(opts.TerminalRequest).To(Equal(options.RequestTTYYes))
					Expect(opts.AppName).To(Equal("app-name"))
					Expect(opts.Command).To(ConsistOf("true"))
				})
			})

			Context("when -c is used more than once", func() {
				BeforeEach(func() {
					args = append(args, "-k", "app-name", "-t", "-c", "echo", "-c", "-n", "-c", "hello!")
				})

				It("handles the app and command correctly", func() {
					Expect(opts.SkipHostValidation).To(BeTrue())
					Expect(opts.TerminalRequest).To(Equal(options.RequestTTYYes))
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
					args = append(args, "-L", "8080:remote:80")
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

})
