package types

import (
	"encoding/json"
	"strconv"

	"code.cloudfoundry.org/bytefmt"
)

// NullByteSizeInMb represents size in a byte format in megabytes.
type NullByteSizeInMb struct {
	IsSet bool

	// Value is a size in MB
	Value uint64
}

func (b NullByteSizeInMb) String() string {
	if !b.IsSet {
		return ""
	}

	return bytefmt.ByteSize(b.Value * bytefmt.MEGABYTE)
}

func (b *NullByteSizeInMb) ParseStringValue(value string) error {
	if value == "" {
		b.IsSet = false
		b.Value = 0
		return nil
	}

	byteSize, fmtErr := bytefmt.ToMegabytes(value)
	if fmtErr != nil {
		return fmtErr
	}

	b.IsSet = true
	b.Value = byteSize

	return nil
}

// ParseUint64Value is used to parse a user provided *uint64 argument.
func (b *NullByteSizeInMb) ParseUint64Value(val *uint64) {
	if val == nil {
		b.IsSet = false
		b.Value = 0
		return
	}

	b.Value = *val
	b.IsSet = true
}

func (b *NullByteSizeInMb) UnmarshalJSON(rawJSON []byte) error {
	var value json.Number
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value.String() == "" {
		b.Value = 0
		b.IsSet = false
		return nil
	}

	valueInt, err := strconv.ParseUint(value.String(), 10, 64)
	if err != nil {
		return err
	}

	b.Value = valueInt
	b.IsSet = true

	return nil
}
