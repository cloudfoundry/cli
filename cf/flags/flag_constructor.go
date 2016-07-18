package flags

func (c *flagContext) NewStringFlag(name string, shortName string, usage string) {
	c.cmdFlags[name] = &StringFlag{Name: name, ShortName: shortName, Usage: usage}
}

func (c *flagContext) NewStringFlagWithDefault(name string, shortName string, usage string, value string) {
	c.cmdFlags[name] = &StringFlag{Name: name, ShortName: shortName, Value: value, Usage: usage}
}

func (c *flagContext) NewBoolFlag(name string, shortName string, usage string) {
	c.cmdFlags[name] = &BoolFlag{Name: name, ShortName: shortName, Usage: usage}
}

func (c *flagContext) NewIntFlag(name string, shortName string, usage string) {
	c.cmdFlags[name] = &IntFlag{Name: name, ShortName: shortName, Usage: usage}
}

func (c *flagContext) NewIntFlagWithDefault(name string, shortName string, usage string, value int) {
	c.cmdFlags[name] = &IntFlag{Name: name, ShortName: shortName, Value: value, Usage: usage}
}

func (c *flagContext) NewFloat64Flag(name string, shortName string, usage string) {
	c.cmdFlags[name] = &Float64Flag{Name: name, ShortName: shortName, Usage: usage}
}

func (c *flagContext) NewFloat64FlagWithDefault(name string, shortName string, usage string, value float64) {
	c.cmdFlags[name] = &Float64Flag{Name: name, ShortName: shortName, Value: value, Usage: usage}
}

func (c *flagContext) NewStringSliceFlag(name string, shortName string, usage string) {
	c.cmdFlags[name] = &StringSliceFlag{Name: name, ShortName: shortName, Usage: usage}
}

func (c *flagContext) NewStringSliceFlagWithDefault(name string, shortName string, usage string, value []string) {
	c.cmdFlags[name] = &StringSliceFlag{Name: name, ShortName: shortName, Value: value, Usage: usage}
}

func (c *flagContext) NewBackwardsCompatibilityFlag(name string, shortName string, usage string) {
	c.cmdFlags["name"] = &BackwardsCompatibilityFlag{}
}
