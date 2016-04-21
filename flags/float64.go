package flags

import "strconv"

type Float64Flag struct {
	Name      string
	Value     float64
	Usage     string
	ShortName string
	Hidden    bool
}

func (f *Float64Flag) Set(v string) {
	i, _ := strconv.ParseFloat(v, 64)
	f.Value = i
}

func (f *Float64Flag) String() string {
	return f.Usage
}

func (f *Float64Flag) GetName() string {
	return f.Name
}

func (f *Float64Flag) GetShortName() string {
	return f.ShortName
}

func (f *Float64Flag) GetValue() interface{} {
	return f.Value
}

func (f *Float64Flag) Visible() bool {
	return !f.Hidden
}
