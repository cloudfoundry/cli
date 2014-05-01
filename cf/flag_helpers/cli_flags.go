package flag_helpers

import (
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

func NewIntFlag(name, usage string) IntFlagWithNoDefault {
	return IntFlagWithNoDefault{cli.IntFlag{Name: name, Usage: usage}}
}

func NewIntFlagWithValue(name, usage string, value int) IntFlagWithNoDefault {
	return IntFlagWithNoDefault{cli.IntFlag{Name: name, Value: value, Usage: usage}}
}

func NewStringFlag(name, usage string) StringFlagWithNoDefault {
	return StringFlagWithNoDefault{cli.StringFlag{Name: name, Usage: usage}}
}

func NewStringSliceFlag(name, usage string) StringSliceFlagWithNoDefault {
	return StringSliceFlagWithNoDefault{cli.StringSliceFlag{Name: name, Usage: usage, Value: &cli.StringSlice{}}}
}

type IntFlagWithNoDefault struct {
	cli.IntFlag
}

type StringFlagWithNoDefault struct {
	cli.StringFlag
}

type StringSliceFlagWithNoDefault struct {
	cli.StringSliceFlag
}

func (f IntFlagWithNoDefault) String() string {
	defaultVal := fmt.Sprintf("'%v'", f.Value)
	return strings.Replace(f.IntFlag.String(), defaultVal, "", 1)
}

func (f StringFlagWithNoDefault) String() string {
	defaultVal := fmt.Sprintf("'%v'", f.Value)
	return strings.Replace(f.StringFlag.String(), defaultVal, "", 1)
}

func (f StringSliceFlagWithNoDefault) String() string {
	return fmt.Sprintf("%s%s \t%s", prefixFor(f.Name), f.Name, f.Usage)
}

func prefixFor(name string) (prefix string) {
	if len(name) == 1 {
		prefix = "-"
	} else {
		prefix = "--"
	}

	return
}
