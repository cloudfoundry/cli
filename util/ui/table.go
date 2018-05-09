package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/lunixbochs/vtclean"
	runewidth "github.com/mattn/go-runewidth"
)

// DefaultTableSpacePadding is the default space padding in tables.
const DefaultTableSpacePadding = 3

// DisplayKeyValueTable outputs a matrix of strings as a table to UI.Out.
// Prefix will be prepended to each row and padding adds the specified number
// of spaces between columns. The final columns may wrap to multiple lines but
// will still be confined to the last column. Wrapping will occur on word
// boundaries.
func (ui *UI) DisplayKeyValueTable(prefix string, table [][]string, padding int) {
	rows := len(table)
	if rows == 0 {
		return
	}

	columns := len(table[0])

	if columns < 2 || !ui.IsTTY {
		ui.DisplayNonWrappingTable(prefix, table, padding)
		return
	}

	ui.displayWrappingTableWithWidth(prefix, table, padding)
}

// DisplayNonWrappingTable outputs a matrix of strings as a table to UI.Out.
// Prefix will be prepended to each row and padding adds the specified number
// of spaces between columns.
func (ui *UI) DisplayNonWrappingTable(prefix string, table [][]string, padding int) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	if len(table) == 0 {
		return
	}

	var columnPadding []int

	rows := len(table)
	columns := len(table[0])
	for col := 0; col < columns; col++ {
		var max int
		for row := 0; row < rows; row++ {
			if strLen := wordSize(table[row][col]); max < strLen {
				max = strLen
			}
		}
		columnPadding = append(columnPadding, max+padding)
	}

	for row := 0; row < rows; row++ {
		fmt.Fprintf(ui.Out, prefix)
		for col := 0; col < columns; col++ {
			data := table[row][col]
			var addedPadding int
			if col+1 != columns {
				addedPadding = columnPadding[col] - wordSize(data)
			}
			fmt.Fprintf(ui.Out, "%s%s", data, strings.Repeat(" ", addedPadding))
		}
		fmt.Fprintf(ui.Out, "\n")
	}
}

// DisplayTableWithHeader outputs a simple non-wrapping table with bolded
// headers.
func (ui *UI) DisplayTableWithHeader(prefix string, table [][]string, padding int) {
	if len(table) == 0 {
		return
	}
	for i, str := range table[0] {
		table[0][i] = ui.modifyColor(str, color.New(color.Bold))
	}

	ui.DisplayNonWrappingTable(prefix, table, padding)
}

func wordSize(str string) int {
	cleanStr := vtclean.Clean(str, false)
	return runewidth.StringWidth(cleanStr)
}
