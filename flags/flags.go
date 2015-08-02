package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudfoundry/cli/flags/flag"
)

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
	StringSlice(string) []string
	IsSet(string) bool
	SkipFlagParsing(bool)
}

type flagContext struct {
	flagsets        map[string]FlagSet
	args            []string
	cmdFlags        map[string]FlagSet //valid flags for command
	cursor          int
	skipFlagParsing bool
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

	c.setDefaultFlagValueIfAny()

	for c.cursor <= len(args)-1 {
		arg := args[c.cursor]

		if !c.skipFlagParsing && (strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--")) {
			flg := strings.TrimLeft(strings.TrimLeft(arg, "-"), "-")

			c.extractEqualSignIfAny(&flg, &args)

			if flagset, ok = c.cmdFlags[flg]; !ok {
				return errors.New("Invalid flag: " + arg)
			}

			switch flagset.GetValue().(type) {
			case bool:
				c.flagsets[flg] = &cliFlags.BoolFlag{Name: flg, Value: c.getBoolFlagValue(args)}
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
			case []string:
				if v, err = c.getFlagValue(args); err != nil {
					return err
				}
				if _, ok = c.flagsets[flg]; !ok {
					c.flagsets[flg] = &cliFlags.StringSliceFlag{Name: flg, Value: []string{v}}
				} else {
					c.flagsets[flg].Set(v)
				}
			}
		} else {
			c.args = append(c.args, args[c.cursor])
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

func (c *flagContext) getBoolFlagValue(args []string) bool {
	if c.cursor >= len(args)-1 {
		return true
	}

	b, err := strconv.ParseBool(args[c.cursor+1])
	if err == nil {
		c.cursor++
		return b
	}
	return true
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

func (c *flagContext) StringSlice(k string) []string {
	if _, ok := c.flagsets[k]; ok {
		v := c.flagsets[k].GetValue()
		switch v.(type) {
		case []string:
			return v.([]string)
		}
	}
	return []string{}
}

func (c *flagContext) SkipFlagParsing(skip bool) {
	c.skipFlagParsing = skip
}

func (c *flagContext) extractEqualSignIfAny(flg *string, args *[]string) {
	if strings.Contains(*flg, "=") {
		tmpAry := strings.SplitN(*flg, "=", 2)
		*flg = tmpAry[0]
		tmpArg := append((*args)[:c.cursor], tmpAry[1])
		*args = append(tmpArg, (*args)[c.cursor:]...)
	}
}

func (c *flagContext) setDefaultFlagValueIfAny() {
	var v interface{}

	for flgName, flg := range c.cmdFlags {
		v = flg.GetValue()
		switch v.(type) {
		case bool:
			if v.(bool) != false {
				c.flagsets[flgName] = &cliFlags.BoolFlag{Name: flgName, Value: v.(bool)}
			}
		case int:
			if v.(int) != 0 {
				c.flagsets[flgName] = &cliFlags.IntFlag{Name: flgName, Value: v.(int)}
			}
		case string:
			if len(v.(string)) != 0 {
				c.flagsets[flgName] = &cliFlags.StringFlag{Name: flgName, Value: v.(string)}
			}
		case []string:
			if len(v.([]string)) != 0 {
				c.flagsets[flgName] = &cliFlags.StringSliceFlag{Name: flgName, Value: v.([]string)}
			}
		}
	}

}
