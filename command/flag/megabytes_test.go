package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Megabytes", func() {
	var megabytes Megabytes

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			megabytes = Megabytes{}
		})

		Context("when the suffix is M", func() {
			It("interprets the number as megabytes", func() {
				err := megabytes.UnmarshalFlag("17M")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytes.Value).To(BeEquivalentTo(17))
			})
		})

		Context("when the suffix is MB", func() {
			It("interprets the number as megabytes", func() {
				err := megabytes.UnmarshalFlag("19MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytes.Value).To(BeEquivalentTo(19))
			})
		})

		Context("when the suffix is G", func() {
			It("interprets the number as gigabytes", func() {
				err := megabytes.UnmarshalFlag("2G")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytes.Value).To(BeEquivalentTo(2048))
			})
		})

		Context("when the suffix is GB", func() {
			It("interprets the number as gigabytes", func() {
				err := megabytes.UnmarshalFlag("3GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytes.Value).To(BeEquivalentTo(3072))
			})
		})

		Context("when the suffix is lowercase", func() {
			It("is case insensitive", func() {
				err := megabytes.UnmarshalFlag("7m")
				Expect(err).ToNot(HaveOccurred())
				Expect(megabytes.Value).To(BeEquivalentTo(7))
			})
		})

		Context("when the megabytes are invalid", func() {
			It("returns an error", func() {
				err := megabytes.UnmarshalFlag("invalid")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		Context("when a decimal is used", func() {
			It("returns an error", func() {
				err := megabytes.UnmarshalFlag("1.2M")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		Context("when the units are missing", func() {
			It("returns an error", func() {
				err := megabytes.UnmarshalFlag("37")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		Context("when the suffix is B", func() {
			It("returns an error", func() {
				err := megabytes.UnmarshalFlag("10B")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		Context("when value is empty", func() {
			It("sets IsSet to false", func() {
				err := megabytes.UnmarshalFlag("")
				Expect(err).NotTo(HaveOccurred())
				Expect(megabytes.IsSet).To(BeFalse())
			})
		})
	})
})
