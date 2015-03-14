package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/flags/flag"
)

/**
 * In main:
 *		- get os args
 *		- new cmdContext obj
 *		- range over os.Args, look for '-'/'--', call flags.Parse() on each of them
 *		- cmdContext obj populated with flags, cmd now has c.Bool(),c.Int(),c.String() etc...
 *		- now populate cmdContext with args, anything w/o '-'/'--', now cmd has c.Args()
 *
 *		now run cmd, pass Context obj over
 * **/

//to do: flag=value (= sign), Args() to return list arguemnts

type FlagSet interface {
	fmt.Stringer
	GetName() string
	GetValue() interface{}
	Set(string)
}

type FlagContext interface {
	Parse(...string) error
	Args() []string
	Int(string) int
	Bool(string) bool
	String(string) string
	IsSet(string) bool
}

type flagContext struct {
	flagsets map[string]FlagSet
	args     []string
	cmdFlags map[string]FlagSet //valid flags for command
	cursor   int
}

func NewFlagContext(cmdFlags map[string]FlagSet) FlagContext {
	return &flagContext{
		flagsets: make(map[string]FlagSet),
		cmdFlags: cmdFlags,
		cursor:   0,
	}
}

func (c *flagContext) Parse(args ...string) error {
	var flagset FlagSet
	var ok bool
	var v string
	var err error

	for c.cursor <= len(args)-1 {
		arg := args[c.cursor]

		if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			flg := strings.TrimLeft(strings.TrimLeft(arg, "-"), "-")
			if flagset, ok = c.cmdFlags[flg]; !ok {
				return errors.New("Invalid flag: " + arg)
			}

			switch flagset.GetValue().(type) {
			case bool:
				c.flagsets[flg] = &cliFlags.BoolFlag{Name: flg, Value: true}
			case int:
				if v, err = c.getFlagValue(args); err != nil {
					return err
				}
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return errors.New("Value for flag '" + flg + "' must be integer")
				}
				c.flagsets[flg] = &cliFlags.IntFlag{Name: flg, Value: int(i)}
			case string:
				if v, err = c.getFlagValue(args); err != nil {
					return err
				}
				c.flagsets[flg] = &cliFlags.StringFlag{Name: flg, Value: v}
			}
		}
		c.cursor++
	}
	return nil
}

func (c *flagContext) getFlagValue(args []string) (string, error) {
	if c.cursor >= len(args)-1 {
		return "", errors.New("No value provided for flag: " + args[c.cursor])
	}

	c.cursor++
	return args[c.cursor], nil
}

func (c *flagContext) Args() []string {
	return c.args
}

func (c *flagContext) IsSet(k string) bool {
	if _, ok := c.flagsets[k]; ok {
		return true
	}
	return false
}

func (c *flagContext) Int(k string) int {
	if _, ok := c.flagsets[k]; ok {
		v := c.flagsets[k].GetValue()
		switch v.(type) {
		case int:
			return v.(int)
		}
	}
	return 0
}

func (c *flagContext) String(k string) string {
	if _, ok := c.flagsets[k]; ok {
		v := c.flagsets[k].GetValue()
		switch v.(type) {
		case string:
			return v.(string)
		}
	}
	return ""
}

func (c *flagContext) Bool(k string) bool {
	if _, ok := c.flagsets[k]; ok {
		v := c.flagsets[k].GetValue()
		switch v.(type) {
		case bool:
			return v.(bool)
		}
	}
	return false
}
