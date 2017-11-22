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
func (n *NullByteSizeInMb) ParseUint64Value(val *uint64) {
	if val == nil {
		n.IsSet = false
		n.Value = 0
		return
	}

	n.Value = *val
	n.IsSet = true
}

func (n *NullByteSizeInMb) UnmarshalJSON(rawJSON []byte) error {
	var value json.Number
	err := json.Unmarshal(rawJSON, &value)
	if err != nil {
		return err
	}

	if value.String() == "" {
		n.Value = 0
		n.IsSet = false
		return nil
	}

	valueInt, err := strconv.ParseUint(value.String(), 10, 64)
	if err != nil {
		return err
	}

	n.Value = valueInt
	n.IsSet = true

	return nil
}
