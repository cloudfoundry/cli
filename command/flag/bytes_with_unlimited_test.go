package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BytesWithUnlimited", func() {
	var bytesWithUnlimited BytesWithUnlimited

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			bytesWithUnlimited = BytesWithUnlimited{}
		})

		When("the suffix is B", func() {
			It("interprets the number as bytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("17B")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(17))
			})
		})

		When("the suffix is K", func() {
			It("interprets the number as kilobytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("64K")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(64 * 1024))
			})
		})

		When("the suffix is KB", func() {
			It("interprets the number as kilobytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("64KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(64 * 1024))
			})
		})

		When("the suffix is M", func() {
			It("interprets the number as megabytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("17M")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(17 * 1024 * 1024))
			})
		})

		When("the suffix is MB", func() {
			It("interprets the number as megabytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("19MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(19 * 1024 * 1024))
			})
		})

		When("the suffix is G", func() {
			It("interprets the number as gigabytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("2G")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(2 * 1024 * 1024 * 1024))
			})
		})

		When("the suffix is GB", func() {
			It("interprets the number as gigabytes", func() {
				err := bytesWithUnlimited.UnmarshalFlag("3GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(3 * 1024 * 1024 * 1024))
			})
		})

		When("the value is -1 (unlimited)", func() {
			It("interprets the number as -1 (unlimited)", func() {
				err := bytesWithUnlimited.UnmarshalFlag("-1")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(-1))
			})

			It("accepts units", func() {
				err := bytesWithUnlimited.UnmarshalFlag("-1T")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(-1))
			})
		})

		When("the value is 0", func() {
			It("sets the value to 0", func() {
				err := bytesWithUnlimited.UnmarshalFlag("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(0))
			})

			It("accepts units", func() {
				err := bytesWithUnlimited.UnmarshalFlag("0TB")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(0))
			})
		})

		When("the suffix is lowercase", func() {
			It("is case insensitive", func() {
				err := bytesWithUnlimited.UnmarshalFlag("7m")
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesWithUnlimited.Value).To(BeEquivalentTo(7 * 1024 * 1024))
			})
		})

		When("the value is invalid", func() {
			It("returns an error", func() {
				err := bytesWithUnlimited.UnmarshalFlag("invalid")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB`,
				}))
			})
		})

		When("a decimal is used", func() {
			It("returns an error", func() {
				err := bytesWithUnlimited.UnmarshalFlag("1.2M")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB`,
				}))
			})
		})

		When("the value is too large", func() {
			It("returns an error", func() {
				err := bytesWithUnlimited.UnmarshalFlag("9999999TB")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB`,
				}))
			})
		})

		When("the units are missing", func() {
			It("returns an error", func() {
				err := bytesWithUnlimited.UnmarshalFlag("37")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB`,
				}))
			})
		})

		When("value is empty", func() {
			It("sets IsSet to false", func() {
				err := bytesWithUnlimited.UnmarshalFlag("")
				Expect(err).NotTo(HaveOccurred())
				Expect(bytesWithUnlimited.IsSet).To(BeFalse())
			})
		})
	})
})
