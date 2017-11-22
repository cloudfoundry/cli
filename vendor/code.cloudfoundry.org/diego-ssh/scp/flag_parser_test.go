package scp_test

import (
	"code.cloudfoundry.org/diego-ssh/scp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FlagParser", func() {
	Describe("ParseFlags", func() {
		Context("when invalid flags are specified", func() {
			It("returns an error", func() {
				_, err := scp.ParseFlags([]string{"scp", "-xxx", "/tmp/foo"})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when unix style command concatenated args are used", func() {
			It("parses command line flags and returns Options", func() {
				scpOptions, err := scp.ParseFlags([]string{"scp", "-tdvprq", "/tmp/foo"})
				Expect(err).NotTo(HaveOccurred())

				Expect(scpOptions.TargetMode).To(BeTrue())
				Expect(scpOptions.SourceMode).To(BeFalse())
				Expect(scpOptions.TargetIsDirectory).To(BeTrue())
				Expect(scpOptions.Verbose).To(BeTrue())
				Expect(scpOptions.PreserveTimesAndMode).To(BeTrue())
				Expect(scpOptions.Recursive).To(BeTrue())
				Expect(scpOptions.Quiet).To(BeTrue())
				Expect(scpOptions.Target).To(Equal("/tmp/foo"))
			})
		})

		Context("when separate flags arguments are used", func() {
			It("parses command line flags and returns Options", func() {
				scpOptions, err := scp.ParseFlags([]string{"scp", "-t", "-d", "-v", "-p", "-r", "/tmp/foo"})
				Expect(err).NotTo(HaveOccurred())

				Expect(scpOptions.TargetMode).To(BeTrue())
				Expect(scpOptions.SourceMode).To(BeFalse())
				Expect(scpOptions.TargetIsDirectory).To(BeTrue())
				Expect(scpOptions.Verbose).To(BeTrue())
				Expect(scpOptions.PreserveTimesAndMode).To(BeTrue())
				Expect(scpOptions.Recursive).To(BeTrue())
				Expect(scpOptions.Target).To(Equal("/tmp/foo"))
			})
		})

		Context("when source mode is specified", func() {
			It("returns Options with SourceMode enabled", func() {
				scpOptions, err := scp.ParseFlags([]string{"scp", "-f", "/tmp/foo"})
				Expect(err).NotTo(HaveOccurred())
				Expect(scpOptions.SourceMode).To(BeTrue())
			})

			It("does not allow TargetMode to be enabled", func() {
				_, err := scp.ParseFlags([]string{"scp", "-ft"})
				Expect(err).To(HaveOccurred())
			})

			Context("Arguments", func() {
				It("populates the Sources with following arguments", func() {
					scpOptions, err := scp.ParseFlags([]string{"scp", "-f", "/foo/bar", "/baz/buzz"})
					Expect(err).NotTo(HaveOccurred())
					Expect(scpOptions.Sources).To(Equal([]string{"/foo/bar", "/baz/buzz"}))
				})

				It("returns an empty string for Target", func() {
					scpOptions, err := scp.ParseFlags([]string{"scp", "-f", "/foo/bar", "/baz/buzz"})
					Expect(err).NotTo(HaveOccurred())
					Expect(scpOptions.Target).To(BeEmpty())
				})

				Context("when no argument is provided", func() {
					It("returns an error", func() {
						_, err := scp.ParseFlags([]string{"scp", "-f"})
						Expect(err).To(MatchError("Must specify at least one source in source mode"))
					})
				})
			})
		})

		Context("when target mode is specified", func() {
			It("returns Options with TargetMode enabled", func() {
				scpOptions, err := scp.ParseFlags([]string{"scp", "-t", "/tmp/foo"})
				Expect(err).NotTo(HaveOccurred())
				Expect(scpOptions.TargetMode).To(BeTrue())
			})

			It("does not allow SourceMode to be enabled", func() {
				_, err := scp.ParseFlags([]string{"scp", "-tf"})
				Expect(err).To(HaveOccurred())
			})

			Context("Arguments", func() {
				It("populates the Target with the argument", func() {
					scpOptions, err := scp.ParseFlags([]string{"scp", "-t", "/foo/bar"})
					Expect(err).NotTo(HaveOccurred())
					Expect(scpOptions.Target).To(Equal("/foo/bar"))
				})

				It("returns an empty array for Sources", func() {
					scpOptions, err := scp.ParseFlags([]string{"scp", "-t", "/foo/bar"})
					Expect(err).NotTo(HaveOccurred())
					Expect(scpOptions.Sources).To(BeEmpty())
				})

				Context("when no argument is provided", func() {
					It("returns an error", func() {
						_, err := scp.ParseFlags([]string{"scp", "-t"})
						Expect(err).To(MatchError("Must specify one target in target mode"))
					})
				})

				Context("when more than one argument is provided", func() {
					It("returns an error", func() {
						_, err := scp.ParseFlags([]string{"scp", "-t", "/foo/bar", "/baz/buzz"})
						Expect(err).To(MatchError("Must specify one target in target mode"))
					})
				})
			})
		})

		Context("when neither target or source mode is specified", func() {
			It("does not allow this", func() {
				_, err := scp.ParseFlags([]string{"scp", ""})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the command is not scp", func() {
			It("returns an error", func() {
				_, err := scp.ParseFlags([]string{"foobar", ""})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ParseCommand", func() {
		var (
			command string
			args    []string
			err     error
		)

		BeforeEach(func() {
			command = "scp -v -f source"
		})

		JustBeforeEach(func() {
			args, err = scp.ParseCommand(command)
		})

		It("returns an string slice from an scp command", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"scp", "-v", "-f", "source"}))
		})

		Context("when the shell lexer returns an error", func() {
			BeforeEach(func() {
				command = "scp -v -f source\\"
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(args).To(BeEmpty())
			})
		})

		Context("when the command string contains escaped spaces as parts of filenames", func() {
			BeforeEach(func() {
				command = "scp -v -f source\\ file"
			})

			It("correctly captures the path as a single argument", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(args).To(Equal([]string{"scp", "-v", "-f", "source file"}))
			})
		})

		Context("when an argument is in quotes", func() {
			BeforeEach(func() {
				command = "scp -v -f \"source\""
			})

			It("correctly captures the path as a single argument", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(args).To(Equal([]string{"scp", "-v", "-f", "source"}))
			})
		})

		Context("when the command contains unexpected whitespace", func() {
			BeforeEach(func() {
				command = "scp -v                   -f                source"
			})

			It("correctly strips excess whitespace", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(args).To(Equal([]string{"scp", "-v", "-f", "source"}))
			})
		})
	})
})
