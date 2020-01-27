package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemoryWithUnlimited", func() {
	var memoryWithUnlimited MemoryWithUnlimited

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			memoryWithUnlimited = MemoryWithUnlimited{}
		})

		When("the suffix is M", func() {
			It("interprets the number as memoryWithUnlimited", func() {
				err := memoryWithUnlimited.UnmarshalFlag("17M")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(17))
			})
		})

		When("the suffix is MB", func() {
			It("interprets the number as memoryWithUnlimited", func() {
				err := memoryWithUnlimited.UnmarshalFlag("19MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(19))
			})
		})

		When("the suffix is G", func() {
			It("interprets the number as gigabytes", func() {
				err := memoryWithUnlimited.UnmarshalFlag("2G")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(2048))
			})
		})

		When("the suffix is GB", func() {
			It("interprets the number as gigabytes", func() {
				err := memoryWithUnlimited.UnmarshalFlag("3GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(3072))
			})
		})

		When("the value is -1 (unlimited)", func() {
			It("interprets the number as memoryWithUnlimited", func() {
				err := memoryWithUnlimited.UnmarshalFlag("-1")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(-1))
			})
		})

		When("the suffix is lowercase", func() {
			It("is case insensitive", func() {
				err := memoryWithUnlimited.UnmarshalFlag("7m")
				Expect(err).ToNot(HaveOccurred())
				Expect(memoryWithUnlimited.Value).To(BeEquivalentTo(7))
			})
		})

		When("the memoryWithUnlimited are invalid", func() {
			It("returns an error", func() {
				err := memoryWithUnlimited.UnmarshalFlag("invalid")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("a decimal is used", func() {
			It("returns an error", func() {
				err := memoryWithUnlimited.UnmarshalFlag("1.2M")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("the units are missing", func() {
			It("returns an error", func() {
				err := memoryWithUnlimited.UnmarshalFlag("37")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("the suffix is B", func() {
			It("returns an error", func() {
				err := memoryWithUnlimited.UnmarshalFlag("10B")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
				}))
			})
		})

		When("value is empty", func() {
			It("sets IsSet to false", func() {
				err := memoryWithUnlimited.UnmarshalFlag("")
				Expect(err).NotTo(HaveOccurred())
				Expect(memoryWithUnlimited.IsSet).To(BeFalse())
			})
		})
	})
})
