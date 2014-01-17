package formatters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByteSize(t *testing.T) {
	assert.Equal(t, ByteSize(100*MEGABYTE), "100M")
	assert.Equal(t, ByteSize(uint64(100.5*MEGABYTE)), "100.5M")
}

func TestParsesByteAmounts(t *testing.T) {
	var (
		megabytes uint64
		err       error
	)

	megabytes, err = ToMegabytes("5M")
	assert.Equal(t, megabytes, uint64(5))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("5m")
	assert.Equal(t, megabytes, uint64(5))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("2G")
	assert.Equal(t, megabytes, uint64(2*1024))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("3T")
	assert.Equal(t, megabytes, uint64(3*1024*1024))
	assert.NoError(t, err)
}

func TestParsesByteAmountsWithLongUnits(t *testing.T) {
	var (
		megabytes uint64
		err       error
	)

	megabytes, err = ToMegabytes("5MB")
	assert.Equal(t, megabytes, uint64(5))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("5mb")
	assert.Equal(t, megabytes, uint64(5))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("2GB")
	assert.Equal(t, megabytes, uint64(2*1024))
	assert.NoError(t, err)

	megabytes, err = ToMegabytes("3TB")
	assert.Equal(t, megabytes, uint64(3*1024*1024))
	assert.NoError(t, err)
}

func TestDoesNotParseAmountsWithoutUnits(t *testing.T) {
	_, err := ToMegabytes("5")
	assert.Error(t, err)
}

func TestDoesNotParseAmountsWithUnknownSuffixes(t *testing.T) {
	_, err := ToMegabytes("5MBB")
	assert.Error(t, err)
}

func TestDoesNotParseAmountsWithUnknownPrefixes(t *testing.T) {
	_, err := ToMegabytes(" 5MB")
	assert.Error(t, err)
}
