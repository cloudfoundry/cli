package table

import (
	"fmt"
	"strings"
	"unicode"
)

const UNKNOWN_HEADER_MAPPING rune = '_'

func NewHeader(title string) Header {
	return Header{
		Key:    KeyifyHeader(title),
		Title:  title,
		Hidden: false,
	}
}

func (t *Table) SetColumnVisibility(headers []Header) error {
	for tableHeaderIdx, _ := range t.Header {
		t.Header[tableHeaderIdx].Hidden = true
	}

	for _, header := range headers {
		foundHeader := false

		for tableHeaderIdx, tableHeader := range t.Header {
			if tableHeader.Key == header.Key || tableHeader.Title == header.Title {
				t.Header[tableHeaderIdx].Hidden = false
				foundHeader = true

				break
			}
		}

		if !foundHeader {
			// key may be empty; if title is present
			return fmt.Errorf("Failed to find header: %s", header.Key)
		}
	}

	return nil
}

func KeyifyHeader(header string) string {
	splittedStrings := strings.Split(cleanHeader(header), " ")
	splittedTrimmedStrings := []string{}
	for _, s := range splittedStrings {
		if s != "" {
			splittedTrimmedStrings = append(splittedTrimmedStrings, s)
		}
	}

	join := strings.Join(splittedTrimmedStrings, "_")
	if len(join) == 0 {
		return string(UNKNOWN_HEADER_MAPPING)
	}
	return join
}

func cleanHeader(header string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		} else {
			return ' '
		}
	}, header)
}
