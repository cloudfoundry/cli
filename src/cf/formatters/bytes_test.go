package formatters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByteSize(t *testing.T) {
	assert.Equal(t, ByteSize(100*MEGABYTE), "100M")
	assert.Equal(t, ByteSize(uint64(100.5*MEGABYTE)), "100.5M")
}
