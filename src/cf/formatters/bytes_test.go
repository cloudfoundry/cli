package formatters_test

import (
	. "cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("5m")
		Expect(megabytes).To(Equal(uint64(5)))
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("2G")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("3T")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		assert.NoError(mr.T(), err)
	})

	It("TestParsesByteAmountsWithLongUnits", func() {
		var (
			megabytes uint64
			err       error
		)

		megabytes, err = ToMegabytes("5MB")
		Expect(megabytes).To(Equal(uint64(5)))
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("5mb")
		Expect(megabytes).To(Equal(uint64(5)))
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("2GB")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		assert.NoError(mr.T(), err)

		megabytes, err = ToMegabytes("3TB")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		assert.NoError(mr.T(), err)
	})

	It("TestDoesNotParseAmountsWithoutUnits", func() {
		_, err := ToMegabytes("5")
		assert.Error(mr.T(), err)
	})

	It("TestDoesNotParseAmountsWithUnknownSuffixes", func() {
		_, err := ToMegabytes("5MBB")
		assert.Error(mr.T(), err)
	})

	It("TestDoesNotParseAmountsWithUnknownPrefixes", func() {
		_, err := ToMegabytes(" 5MB")
		assert.Error(mr.T(), err)
	})
})
