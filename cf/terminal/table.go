package terminal

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Table interface {
	Add(row ...string)
	Print()
}

type PrintableTable struct {
	ui                  UI
	headers             []string
	headerPrinted       bool
	maxRuneCountLengths []int
	maxStringLengths    []int
	rows                [][]string
}

func NewTable(ui UI, headers []string) Table {
	return &PrintableTable{
		ui:                  ui,
		headers:             headers,
		maxRuneCountLengths: make([]int, len(headers)),
		maxStringLengths:    make([]int, len(headers)),
	}
}

func (t *PrintableTable) Add(row ...string) {
	t.rows = append(t.rows, row)
}

func (t *PrintableTable) Print() {
	for _, row := range append(t.rows, t.headers) {
		t.calculateMaxSize(row)
	}

	if t.headerPrinted == false {
		t.printHeader()
		t.headerPrinted = true
	}

	for _, line := range t.rows {
		t.printRow(line)
	}

	t.rows = [][]string{}
}

func (t *PrintableTable) calculateMaxSize(row []string) {
	for index, value := range row {
		runeCount := utf8.RuneCountInString(Decolorize(value))
		stringLength := len(Decolorize(value))

		if t.maxRuneCountLengths[index] < runeCount {
			t.maxRuneCountLengths[index] = runeCount
		}

		if t.maxStringLengths[index] < stringLength {
			t.maxStringLengths[index] = stringLength
		}
	}
}

func (t *PrintableTable) printHeader() {
	output := ""
	for col, value := range t.headers {
		output = output + t.cellValue(col, HeaderColor(value))
	}
	t.ui.Say(output)
}

func (t *PrintableTable) printRow(row []string) {
	output := ""
	for columnIndex, value := range row {
		if columnIndex == 0 {
			value = TableContentHeaderColor(value)
		}

		output = output + t.cellValue(columnIndex, value)
	}
	t.ui.Say("%s", output)
}

func (t *PrintableTable) cellValue(col int, value string) string {
	padding := ""

	if col < len(t.headers)-1 {
		var count int

		if utf8.RuneCountInString(value) == len(value) {
			if t.maxRuneCountLengths[col] == t.maxStringLengths[col] {
				count = t.maxRuneCountLengths[col] - utf8.RuneCountInString(Decolorize(value))
			} else {
				count = t.maxRuneCountLengths[col] - len(Decolorize(value))
			}
		} else {
			if t.maxRuneCountLengths[col] == t.maxStringLengths[col] {
				count = t.maxRuneCountLengths[col] - len(Decolorize(value)) + utf8.RuneCountInString(Decolorize(value))
			} else {
				count = t.maxStringLengths[col] - len(Decolorize(value))
			}
		}

		padding = strings.Repeat(" ", count)
	}

	return fmt.Sprintf("%s%s   ", value, padding)
}
