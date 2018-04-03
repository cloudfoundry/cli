package table

import (
	"time"

	semver "github.com/cppforlife/go-semi-semantic/version"
)

type Table struct {
	Title   string
	Content string

	Header []Header

	SortBy []ColumnSort

	// Either sections or rows should be provided
	Sections []Section
	Rows     [][]Value

	Notes []string

	// Formatting
	HeaderFormatFunc func(string, ...interface{}) string
	DataOnly         bool
	FillFirstColumn  bool
	BackgroundStr    string
	BorderStr        string
	Transpose        bool
}

type Header struct {
	Key    string
	Title  string
	Hidden bool
}

type Section struct {
	FirstColumn Value
	Rows        [][]Value
}

type ColumnSort struct {
	Column int
	Asc    bool
}

type Value interface {
	Value() Value
	String() string
	Compare(Value) int
}

type ValueString struct {
	S string
}

type EmptyValue struct{}

type ValueStrings struct {
	S []string
}

type ValueInt struct {
	I int
}

type ValueBytes struct {
	I uint64
}

type ValueTime struct {
	T time.Time
}

type ValueBool struct {
	B bool
}

type ValueVersion struct {
	V semver.Version
}

type ValueInterface struct {
	I interface{}
}

type ValueError struct {
	E error
}

type ValueNone struct{}

type ValueFmt struct {
	V     Value
	Error bool
	Func  func(string, ...interface{}) string
}

type ValueSuffix struct {
	V      Value
	Suffix string
}
