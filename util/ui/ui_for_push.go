package ui

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/types"

	"github.com/fatih/color"
)

var ErrValueMissmatch = errors.New("values provided were of different types")

type Change struct {
	Header       string
	CurrentValue interface{}
	NewValue     interface{}
	HiddenValue  bool
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
		err := ui.DisplayChangeForPush(change.Header, padding, change.HiddenValue, change.CurrentValue, change.NewValue)
		if err != nil {
			return err
		}
	}

	return nil
}

// DisplayChangeForPush will display the header and old/new value with the
// appropriately red/green minuses and pluses.
func (ui *UI) DisplayChangeForPush(header string, stringTypePadding int, hiddenValue bool, originalValue interface{}, newValue interface{}) error {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	originalType := reflect.ValueOf(originalValue).Type()
	newType := reflect.ValueOf(newValue).Type()
	if originalType != newType {
		return ErrValueMissmatch
	}

	offset := strings.Repeat(" ", stringTypePadding)

	switch oVal := originalValue.(type) {
	case int:
		nVal := newValue.(int)
		ui.displayDiffForInt(offset, header, oVal, nVal)
	case types.NullInt:
		nVal := newValue.(types.NullInt)
		ui.displayDiffForNullInt(offset, header, oVal, nVal)
	case string:
		nVal := newValue.(string)
		ui.displayDiffForString(offset, header, hiddenValue, oVal, nVal)
	case []string:
		nVal := newValue.([]string)
		if len(oVal) == 0 && len(nVal) == 0 {
			return nil
		}

		ui.displayDiffForStrings(offset, header, oVal, nVal)
	case map[string]string:
		nVal := newValue.(map[string]string)
		if len(oVal) == 0 && len(nVal) == 0 {
			return nil
		}

		ui.displayDiffForMapStringString(offset, header, oVal, nVal)
	default:
		panic(fmt.Sprintf("diff display does not have case for '%s'", header))
	}
	return nil
}

func (ui UI) displayDiffForInt(offset string, header string, oldValue int, newValue int) {
	if oldValue != newValue {
		formattedOld := fmt.Sprintf("- %s%s%d", ui.TranslateText(header), offset, oldValue)
		formattedNew := fmt.Sprintf("+ %s%s%d", ui.TranslateText(header), offset, newValue)

		if oldValue != 0 {
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
		}
		fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
	} else {
		fmt.Fprintf(ui.Out, "  %s%s%d\n", ui.TranslateText(header), offset, oldValue)
	}
}

func (ui UI) displayDiffForMapStringString(offset string, header string, oldMap map[string]string, newMap map[string]string) {
	var oldKeys []string
	for key := range oldMap {
		oldKeys = append(oldKeys, key)
	}

	var newKeys []string
	for key := range newMap {
		newKeys = append(newKeys, key)
	}

	sortedKeys := sortedUniqueArray(oldKeys, newKeys)

	fmt.Fprintf(ui.Out, "  %s\n", ui.TranslateText(header))
	for _, key := range sortedKeys {
		newVal, ok := newMap[key]
		if !ok {
			formattedOld := fmt.Sprintf("-   %s", key)
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
			continue
		}
		oldVal, ok := oldMap[key]
		if !ok {
			formattedNew := fmt.Sprintf("+   %s", key)
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
			continue
		}

		if oldVal == newVal {
			fmt.Fprintf(ui.Out, "    %s\n", key)
		} else {
			formattedOld := fmt.Sprintf("-   %s", key)
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
			formattedNew := fmt.Sprintf("+   %s", key)
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
		}
	}
}

func (ui UI) displayDiffForNullInt(offset string, header string, oldValue types.NullInt, newValue types.NullInt) {
	if oldValue != newValue {
		formattedOld := fmt.Sprintf("- %s%s%d", ui.TranslateText(header), offset, oldValue.Value)
		formattedNew := fmt.Sprintf("+ %s%s%d", ui.TranslateText(header), offset, newValue.Value)

		if oldValue.IsSet {
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
		}
		fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
	} else {
		fmt.Fprintf(ui.Out, "  %s%s%d\n", ui.TranslateText(header), offset, oldValue.Value)
	}
}

func (ui UI) displayDiffForString(offset string, header string, hiddenValue bool, oVal string, nVal string) {
	if oVal != nVal {
		var formattedOld, formattedNew string
		if hiddenValue {
			formattedOld = fmt.Sprintf("- %s%s%s", ui.TranslateText(header), offset, RedactedValue)
			formattedNew = fmt.Sprintf("+ %s%s%s", ui.TranslateText(header), offset, RedactedValue)
		} else {
			formattedOld = fmt.Sprintf("- %s%s%s", ui.TranslateText(header), offset, oVal)
			formattedNew = fmt.Sprintf("+ %s%s%s", ui.TranslateText(header), offset, nVal)
		}

		if oVal != "" {
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedOld, color.New(color.FgRed)))
		}
		if nVal != "" {
			fmt.Fprintln(ui.Out, ui.modifyColor(formattedNew, color.New(color.FgGreen)))
		}
	} else {
		if hiddenValue {
			fmt.Fprintf(ui.Out, "  %s%s%s\n", ui.TranslateText(header), offset, RedactedValue)
		} else {
			fmt.Fprintf(ui.Out, "  %s%s%s\n", ui.TranslateText(header), offset, oVal)
		}
	}
}

func (ui UI) displayDiffForStrings(offset string, header string, oldList []string, newList []string) {
	fmt.Fprintf(ui.Out, "  %s\n", ui.TranslateText(header))

	fullList := sortedUniqueArray(oldList, newList)
	for _, item := range fullList {
		inOld := existsIn(item, oldList)
		inNew := existsIn(item, newList)

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
