package formatters

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func ByteSize(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

func BytesFromString(s string) (bytes uint64, err error) {
	unit := string(s[len(s)-1])
	stringValue := s[0 : len(s)-1]

	value, err := strconv.ParseUint(stringValue, 10, 0)
	if err != nil {
		return
	}

	switch unit {
	case "T":
		bytes = value * TERABYTE
	case "G":
		bytes = value * GIGABYTE
	case "M":
		bytes = value * MEGABYTE
	case "K":
		bytes = value * KILOBYTE
	}

	if bytes == 0 {
		err = errors.New("Could not parse byte string")
	}

	return
}
