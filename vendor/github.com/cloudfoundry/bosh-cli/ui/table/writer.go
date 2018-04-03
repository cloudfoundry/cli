package table

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Writer struct {
	w         io.Writer
	emptyStr  string
	bgStr     string
	borderStr string

	rows   []writerRow
	widths map[int]int
}

type writerCell struct {
	Value   Value
	String  string
	IsEmpty bool
}

type writerRow struct {
	Values   []writerCell
	IsSpacer bool
}

type hasCustomWriter interface {
	Fprintf(io.Writer, string, ...interface{}) (int, error)
}

func NewWriter(w io.Writer, emptyStr, bgStr, borderStr string) *Writer {
	return &Writer{
		w:         w,
		emptyStr:  emptyStr,
		bgStr:     bgStr,
		borderStr: borderStr,
		widths:    map[int]int{},
	}
}

func (w *Writer) Write(headers []Header, vals []Value) {
	rowsToAdd := 1
	colsWithRows := [][]writerCell{}

	visibleHeaderIndex := 0
	for i, val := range vals {
		if len(headers) > 0 && headers[i].Hidden {
			continue
		}

		var rowsInCol []writerCell

		cleanStr := strings.Replace(val.String(), "\r", "", -1)
		lines := strings.Split(cleanStr, "\n")

		if len(lines) == 1 && lines[0] == "" {
			cell := writerCell{Value: val, String: w.emptyStr}

			if reflect.TypeOf(val) == reflect.TypeOf(EmptyValue{}) {
				cell.IsEmpty = true
			}
			rowsInCol = append(rowsInCol, cell)
		} else {
			for _, line := range lines {
				cell := writerCell{Value: val, String: line}
				if reflect.TypeOf(val) == reflect.TypeOf(EmptyValue{}) {
					cell.IsEmpty = true
				}
				rowsInCol = append(rowsInCol, cell)
			}
		}

		rowsInColLen := len(rowsInCol)

		for _, cell := range rowsInCol {
			if len(cell.String) > w.widths[visibleHeaderIndex] {
				w.widths[visibleHeaderIndex] = len(cell.String)
			}
		}

		colsWithRows = append(colsWithRows, rowsInCol)

		if rowsInColLen > rowsToAdd {
			rowsToAdd = rowsInColLen
		}

		visibleHeaderIndex++
	}

	for i := 0; i < rowsToAdd; i++ {
		var row writerRow

		rowIsSeparator := true
		for _, col := range colsWithRows {
			if i < len(col) {
				row.Values = append(row.Values, col[i])
				if !col[i].IsEmpty {
					rowIsSeparator = false
				}
			} else {
				row.Values = append(row.Values, writerCell{})
			}
		}
		row.IsSpacer = rowIsSeparator

		w.rows = append(w.rows, row)
	}
}

func (w *Writer) Flush() error {
	for _, row := range w.rows {
		if row.IsSpacer {
			_, err := fmt.Fprintln(w.w)
			if err != nil {
				return err
			}
			continue
		}

		lastColIdx := len(row.Values) - 1
		for colIdx, col := range row.Values {
			if customWriter, ok := col.Value.(hasCustomWriter); ok {
				_, err := customWriter.Fprintf(w.w, "%s", col.String)
				if err != nil {
					return err
				}
			} else {
				_, err := fmt.Fprintf(w.w, "%s", col.String)
				if err != nil {
					return err
				}
			}

			paddingSize := w.widths[colIdx] - len(col.String)
			if colIdx == lastColIdx {
				_, err := fmt.Fprintf(w.w, w.borderStr)
				if err != nil {
					return err
				}
			} else {
				_, err := fmt.Fprintf(w.w, strings.Repeat(w.bgStr, paddingSize)+w.borderStr)
				if err != nil {
					return err
				}
			}
		}

		_, err := fmt.Fprintln(w.w)
		if err != nil {
			return err
		}
	}

	return nil
}
