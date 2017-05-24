package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

var ErrValueMissmatch = errors.New("values provided were of different types")

type Change struct {
	Header       string
	CurrentValue interface{}
	NewValue     interface{}
}

// DisplayChangesForPush will display the set of changes via
// DisplayChangeForPush in the order given.
func (ui *UI) DisplayChangesForPush(changeSet []Change) error {
	if len(changeSet) == 0 {
		return nil
	}

	var columnWidth int
	for _, change := range changeSet {
		if width := wordSize(ui.TranslateText(change.Header)); width > columnWidth {
			columnWidth = width
		}
	}

	for _, change := range changeSet {
		padding := columnWidth - wordSize(ui.TranslateText(change.Header)) + 3
		err := ui.DisplayChangeForPush(change.Header, padding, change.CurrentValue, change.NewValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// DisplayChangeForPush will display the header and old/new value with the
// appropriately red/green minuses and pluses.
func (ui *UI) DisplayChangeForPush(header string, stringTypePadding int, originalValue interface{}, newValue interface{}) error {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	switch oVal := originalValue.(type) {
	case string:
		nVal, ok := newValue.(string)
		if !ok {
			return ErrValueMissmatch
		}

		offset := strings.Repeat(" ", stringTypePadding)

		if oVal != nVal {
			formattedOld := fmt.Sprintf("- %s%s%s", ui.TranslateText(header), offset, oVal)
			formattedNew := fmt.Sprintf("+ %s%s%s", ui.TranslateText(header), offset, nVal)

			if oVal != "" {
				fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
			}
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
		} else {
			fmt.Fprintf(ui.Out, "  %s%s%s\n", ui.TranslateText(header), offset, oVal)
		}
	case []string:
		nVal, ok := newValue.([]string)
		if !ok {
			return ErrValueMissmatch
		}

		fmt.Fprintf(ui.Out, "  %s\n", ui.TranslateText(header))

		fullList := sortedUniqueArray(oVal, nVal)
		for _, item := range fullList {
			inOld := existsIn(item, oVal)
			inNew := existsIn(item, nVal)

			if inOld && inNew {
				fmt.Fprintf(ui.Out, "    %s\n", item)
			} else if inOld {
				formattedOld := fmt.Sprintf("-   %s", item)
				fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
			} else {
				formattedNew := fmt.Sprintf("+   %s", item)
				fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
			}
		}
	}
	return nil
}

func existsIn(str string, ary []string) bool {
	for _, val := range ary {
		if val == str {
			return true
		}
	}
	return false
}

func sortedUniqueArray(ary1 []string, ary2 []string) []string {
	uniq := append([]string{}, ary1...)

	for _, str := range ary2 {
		if !existsIn(str, uniq) {
			uniq = append(uniq, str)
		}
	}

	sort.Strings(uniq)
	return uniq
}
