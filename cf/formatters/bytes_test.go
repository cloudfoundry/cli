package formatters_test

import (
	. "code.cloudfoundry.org/cli/cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("formatting bytes to / from strings", func() {
	Describe("ByteSize()", func() {
		It("converts bytes to a human readable description", func() {
			Expect(ByteSize(100 * MEGABYTE)).To(Equal("100M"))
			Expect(ByteSize(100 * GIGABYTE)).To(Equal("100G"))
			Expect(ByteSize(int64(100.5 * MEGABYTE))).To(Equal("100.5M"))
			Expect(ByteSize(int64(50))).To(Equal("50B"))
		})

		It("returns 0 byte as '0' without any unit", func() {
			Expect(ByteSize(int64(0))).To(Equal("0"))
		})
	})

	It("parses byte amounts with short units (e.g. M, G)", func() {
		var (
			megabytes int64
			err       error
		)

		megabytes, err = ToMegabytes("5M")
		Expect(megabytes).To(Equal(int64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5m")
		Expect(megabytes).To(Equal(int64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2G")
		Expect(megabytes).To(Equal(int64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3T")
		Expect(megabytes).To(Equal(int64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("parses byte amounts with long units (e.g MB, GB)", func() {
		var (
			megabytes int64
			err       error
		)

		megabytes, err = ToMegabytes("5MB")
		Expect(megabytes).To(Equal(int64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5mb")
		Expect(megabytes).To(Equal(int64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2GB")
		Expect(megabytes).To(Equal(int64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3TB")
		Expect(megabytes).To(Equal(int64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error when the unit is missing", func() {
		_, err := ToMegabytes("5")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})

	It("returns an error when the unit is unrecognized", func() {
		_, err := ToMegabytes("5MBB")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})

	It("allows whitespace before and after the value", func() {
		megabytes, err := ToMegabytes("\t\n\r 5MB ")
		Expect(megabytes).To(Equal(int64(5)))
		Expect(err).NotTo(HaveOccurred())
	})
})
