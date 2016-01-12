package terminal

import (
	"fmt"
	"strings"
)

type Table interface {
	Add(row ...string)
	Print()
}

type PrintableTable struct {
	ui              UI
	headers         []string
	headerPrinted   bool
	maxValueLengths []int
	rows            [][]string
}

func NewTable(ui UI, headers []string) Table {
	return &PrintableTable{
		ui:              ui,
		headers:         headers,
		maxValueLengths: make([]int, len(headers)),
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
		l := visibleSize(Decolorize(value))
		if t.maxValueLengths[index] < l {
			t.maxValueLengths[index] = l
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
	maxVisibleSize := t.maxValueLengths[col]

	if col < len(t.headers)-1 {
		thisVisibleSize := visibleSize(Decolorize(value))
		padding = strings.Repeat(` `, maxVisibleSize-thisVisibleSize)
	}

	return fmt.Sprintf("%s%s   ", value, padding)
}

func visibleSize(s string) int {
	r := strings.NewReader(s)

	var size int
	for range s {
		_, runeSize, err := r.ReadRune()
		if err != nil {
			panic(fmt.Sprintf("error when calculating visible size of: %s", s))
		}

		if runeSize == 3 {
			size += 2 // Kanji and Katakana characters appear as double-width
		} else {
			size += 1
		}
	}

	return size
}
