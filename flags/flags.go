package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FlagSet interface {
	fmt.Stringer
	GetName() string
	GetShortName() string
	GetValue() interface{}
	Set(string)
	Visible() bool
}

type FlagContext interface {
	Parse(...string) error
	Args() []string
	Int(string) int
	Float64(string) float64
	Bool(string) bool
	String(string) string
	StringSlice(string) []string
	IsSet(string) bool
	SkipFlagParsing(bool)
	NewStringFlag(name string, shortName string, usage string)
	NewStringFlagWithDefault(name string, shortName string, usage string, value string)
	NewBoolFlag(name string, shortName string, usage string)
	NewIntFlag(name string, shortName string, usage string)
	NewIntFlagWithDefault(name string, shortName string, usage string, value int)
	NewFloat64Flag(name string, shortName string, usage string)
	NewFloat64FlagWithDefault(name string, shortName string, usage string, value float64)
	NewStringSliceFlag(name string, shortName string, usage string)
	NewStringSliceFlagWithDefault(name string, shortName string, usage string, value []string)
	ShowUsage(leadingSpace int) string
}

type flagContext struct {
	flagsets        map[string]FlagSet
	args            []string
	cmdFlags        map[string]FlagSet //valid flags for command
	cursor          int
	skipFlagParsing bool
}

func New() FlagContext {
	return &flagContext{
		flagsets: make(map[string]FlagSet),
		cmdFlags: make(map[string]FlagSet),
		cursor:   0,
	}
}

func NewFlagContext(cmdFlags map[string]FlagSet) FlagContext {
	return &flagContext{
		flagsets: make(map[string]FlagSet),
		cmdFlags: cmdFlags,
		cursor:   0,
	}
}

func (c *flagContext) Parse(args ...string) error {
	c.setDefaultFlagValueIfAny()

	for c.cursor <= len(args)-1 {
		arg := args[c.cursor]

		if !c.skipFlagParsing && (strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--")) {
			flg := strings.TrimLeft(strings.TrimLeft(arg, "-"), "-")

			c.extractEqualSignIfAny(&flg, &args)

			flagset, ok := c.cmdFlags[flg]
			if !ok {
				flg = c.getFlagNameWithShortName(flg)
				if flagset, ok = c.cmdFlags[flg]; !ok {
					return errors.New("Invalid flag: " + arg)
				}
			}

			switch flagset.GetValue().(type) {
			case bool:
				c.flagsets[flg] = &BoolFlag{Name: flg, Value: c.getBoolFlagValue(args)}
			case int:
				v, err := c.getFlagValue(args)
				if err != nil {
					return err
				}
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return errors.New("Value for flag '" + flg + "' must be an integer")
				}
				c.flagsets[flg] = &IntFlag{Name: flg, Value: int(i)}
			case float64:
				v, err := c.getFlagValue(args)
				if err != nil {
					return err
				}
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return errors.New("Value for flag '" + flg + "' must be a float64")
				}
				c.flagsets[flg] = &Float64Flag{Name: flg, Value: float64(i)}
			case string:
				v, err := c.getFlagValue(args)
				if err != nil {
					return err
				}
				c.flagsets[flg] = &StringFlag{Name: flg, Value: v}
			case []string:
				v, err := c.getFlagValue(args)
				if err != nil {
					return err
				}
				if _, ok = c.flagsets[flg]; !ok {
					c.flagsets[flg] = &StringSliceFlag{Name: flg, Value: []string{v}}
				} else {
					c.flagsets[flg].Set(v)
				}
			case backwardsCompatibilityType:
				// do nothing
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
	return c.isFlagProvided(&k)
}

func (c *flagContext) Int(k string) int {
	if !c.isFlagProvided(&k) {
		return 0
	}

	v := c.flagsets[k].GetValue()
	switch v.(type) {
	case int:
		return v.(int)
	}

	return 0
}

func (c *flagContext) Float64(k string) float64 {
	if !c.isFlagProvided(&k) {
		return 0
	}

	v := c.flagsets[k].GetValue()
	switch v.(type) {
	case float64:
		return v.(float64)
	}
	return 0
}

func (c *flagContext) String(k string) string {
	if !c.isFlagProvided(&k) {
		return ""
	}

	v := c.flagsets[k].GetValue()
	switch v.(type) {
	case string:
		return v.(string)
	}
	return ""
}

func (c *flagContext) Bool(k string) bool {
	if !c.isFlagProvided(&k) {
		return false
	}

	v := c.flagsets[k].GetValue()
	switch v.(type) {
	case bool:
		return v.(bool)
	}

	return false
}

func (c *flagContext) StringSlice(k string) []string {
	if !c.isFlagProvided(&k) {
		return []string{}
	}

	v := c.flagsets[k].GetValue()
	switch v.(type) {
	case []string:
		return v.([]string)
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
				c.flagsets[flgName] = &BoolFlag{Name: flgName, Value: v.(bool)}
			}
		case int:
			if v.(int) != 0 {
				c.flagsets[flgName] = &IntFlag{Name: flgName, Value: v.(int)}
			}
		case float64:
			if v.(float64) != 0 {
				c.flagsets[flgName] = &Float64Flag{Name: flgName, Value: v.(float64)}
			}
		case string:
			if len(v.(string)) != 0 {
				c.flagsets[flgName] = &StringFlag{Name: flgName, Value: v.(string)}
			}
		case []string:
			if len(v.([]string)) != 0 {
				c.flagsets[flgName] = &StringSliceFlag{Name: flgName, Value: v.([]string)}
			}
		}
	}

}

func (c *flagContext) getFlagNameWithShortName(shortName string) string {
	for n, f := range c.cmdFlags {
		if f.GetShortName() == shortName {
			return n
		}
	}
	return ""
}

func (c *flagContext) isFlagProvided(flg *string) bool {
	if _, ok := c.flagsets[*flg]; !ok {
		*flg = c.getFlagNameWithShortName(*flg)
		if _, ok := c.flagsets[*flg]; !ok {
			return false
		}
	}

	return true
}
