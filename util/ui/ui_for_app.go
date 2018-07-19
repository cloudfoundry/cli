package ui

import (
	"strings"

	"github.com/fatih/color"
)

func (ui *UI) DisplayInstancesTableForApp(table [][]string) {
	redColor := color.New(color.FgRed, color.Bold)
	trDown, trCrashed := ui.TranslateText("down"), ui.TranslateText("crashed")

	for i, row := range table {
		if row[1] == trDown || row[1] == trCrashed {
			table[i][1] = ui.modifyColor(row[1], redColor)
		}
	}
	ui.DisplayTableWithHeader("", table, DefaultTableSpacePadding)
}

func (ui *UI) DisplayKeyValueTableForApp(table [][]string) {
	runningInstances := strings.Split(table[2][1], "/")[0]
	state := table[1][1]

	if runningInstances == "0" && state != ui.TranslateText("stopped") {
		redColor := color.New(color.FgRed, color.Bold)
		table[1][1] = ui.modifyColor(table[1][1], redColor)
		table[2][1] = ui.modifyColor(table[2][1], redColor)
	}
	ui.DisplayKeyValueTable("", table, 3)
}
