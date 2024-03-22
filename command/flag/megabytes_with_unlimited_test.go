package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MegabytesWithUnlimited", func() {
	var megabytesWithUnlimited MegabytesWithUnlimited

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			megabytesWithUnlimited = MegabytesWithUnlimited{}
		})

		When("the suffix is M", func() {
			It("interprets the number as megabytesWithUnlimited", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("17M")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(17))
			})
		})

		When("the suffix is MB", func() {
			It("interprets the number as megabytesWithUnlimited", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("19MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(19))
			})
		})

		When("the suffix is G", func() {
			It("interprets the number as gigabytes", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("2G")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(2048))
			})
		})

		When("the suffix is GB", func() {
			It("interprets the number as gigabytes", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("3GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(3072))
			})
		})

		When("the value is -1 (unlimited)", func() {
			It("interprets the number as megabytesWithUnlimited", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("-1")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(-1))
			})
		})

		When("the suffix is lowercase", func() {
			It("is case insensitive", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("7m")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytesWithUnlimited.Value).To(BeEquivalentTo(7))
			})
		})

		When("the megabytesWithUnlimited are invalid", func() {
			It("returns an error", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("invalid")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("a decimal is used", func() {
			It("returns an error", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("1.2M")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("the units are missing", func() {
			It("returns an error", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("37")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("the suffix is B", func() {
			It("returns an error", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("10B")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("value is empty", func() {
			It("sets IsSet to false", func() {
				err := megabytesWithUnlimited.UnmarshalFlag("")
				Expect(err).NotTo(HaveOccurred())
				Expect(megabytesWithUnlimited.IsSet).To(BeFalse())
			})
		})
	})
})
