package formatters

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestByteSize", func() {

			assert.Equal(mr.T(), ByteSize(100*MEGABYTE), "100M")
			assert.Equal(mr.T(), ByteSize(uint64(100.5*MEGABYTE)), "100.5M")
		})
		It("TestParsesByteAmounts", func() {

			var (
				megabytes uint64
				err       error
			)

			megabytes, err = ToMegabytes("5M")
			assert.Equal(mr.T(), megabytes, uint64(5))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("5m")
			assert.Equal(mr.T(), megabytes, uint64(5))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("2G")
			assert.Equal(mr.T(), megabytes, uint64(2*1024))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("3T")
			assert.Equal(mr.T(), megabytes, uint64(3*1024*1024))
			assert.NoError(mr.T(), err)
		})
		It("TestParsesByteAmountsWithLongUnits", func() {

			var (
				megabytes uint64
				err       error
			)

			megabytes, err = ToMegabytes("5MB")
			assert.Equal(mr.T(), megabytes, uint64(5))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("5mb")
			assert.Equal(mr.T(), megabytes, uint64(5))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("2GB")
			assert.Equal(mr.T(), megabytes, uint64(2*1024))
			assert.NoError(mr.T(), err)

			megabytes, err = ToMegabytes("3TB")
			assert.Equal(mr.T(), megabytes, uint64(3*1024*1024))
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
}
