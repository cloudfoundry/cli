package formatters

import (
	"errors"
	"fmt"
	"regexp"
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

var bytesPattern *regexp.Regexp = regexp.MustCompile("(?i)^(-?\\d+)([KMGT])B?$")

func ToMegabytes(s string) (megabytes uint64, err error) {
	parts := bytesPattern.FindStringSubmatch(s)
	if len(parts) < 3 {
		err = errors.New("Could not parse byte quantity '" + s + "'")
		return
	}

	var quantity uint64
	value, err := strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return
	}

	if value < 1 {
		err = errors.New("Byte quantity must be a positive integer with a unit of measurement like M, MB, G, or GB")
		return
	} else {
		quantity = uint64(value)
	}

	var bytes uint64
	unit := strings.ToUpper(parts[2])
	switch unit {
	case "T":
		bytes = quantity * TERABYTE
	case "G":
		bytes = quantity * GIGABYTE
	case "M":
		bytes = quantity * MEGABYTE
	case "K":
		bytes = quantity * KILOBYTE
	}

	megabytes = bytes / MEGABYTE
	return
}
