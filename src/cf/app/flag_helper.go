package app

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

type IntFlagWithNoDefault struct {
	cli.IntFlag
}

func (f IntFlagWithNoDefault) String() string {
	defaultVal := fmt.Sprintf("'%v'", f.Value)
	return strings.Replace(f.IntFlag.String(), defaultVal, "", 1)
}

type StringFlagWithNoDefault struct {
	cli.StringFlag
}

func (f StringFlagWithNoDefault) String() string {
	defaultVal := fmt.Sprintf("'%v'", f.Value)
	return strings.Replace(f.StringFlag.String(), defaultVal, "", 1)
}
