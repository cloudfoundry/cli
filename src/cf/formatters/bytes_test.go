package formatters_test

import (
	. "cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestByteSize", func() {
		Expect(ByteSize(100 * MEGABYTE)).To(Equal("100M"))
		Expect(ByteSize(uint64(100.5 * MEGABYTE))).To(Equal("100.5M"))
	})

	It("TestParsesByteAmounts", func() {
		var (
			megabytes uint64
			err       error
		)

		megabytes, err = ToMegabytes("5M")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5m")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2G")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3T")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("TestParsesByteAmountsWithLongUnits", func() {
		var (
			megabytes uint64
			err       error
		)

		megabytes, err = ToMegabytes("5MB")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5mb")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2GB")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3TB")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("TestDoesNotParseAmountsWithoutUnits", func() {
		_, err := ToMegabytes("5")
		Expect(err).To(HaveOccurred())
	})

	It("TestDoesNotParseAmountsWithUnknownSuffixes", func() {
		_, err := ToMegabytes("5MBB")
		Expect(err).To(HaveOccurred())
	})

	It("TestDoesNotParseAmountsWithUnknownPrefixes", func() {
		_, err := ToMegabytes(" 5MB")
		Expect(err).To(HaveOccurred())
	})

	It("Does not allow non-positive values", func() {
		_, err := ToMegabytes("-5MB")
		Expect(err).To(HaveOccurred())
	})
})
