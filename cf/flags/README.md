# flags - Golang command-line flag parser
[![GoDoc](https://godoc.org/github.com/cloudfoundry/cli/cf/flags?status.svg)](https://godoc.org/github.com/cloudfoundry/cli/cf/flags)

- Fully tested, reliable
- Support flag ShortName (Alias)
- Catches any non-defined flags, and any invalid flag values
- Flags can come before or after the arguments. The followings are all valid inputs:
```bash
$ testapp -i 100 -m 500 arg1 arg2   # flags go first
$ testapp arg1 arg2 --i 100 -m 500  # flags go last
$ testapp arg1 -i 100 arg2 -m=500   # flags go in between arguments
```
The parsed results for all 3 statements are identical: `i=100`, `Args=[arg1, arg2]`, `m=500`

# Installation
```bash
go get github.com/cloudfoundry/cli/cf/flags  # installs the flags library
```

# Usage
```Go
package main

import "github.com/cloudfoundry/cli/cf/flags"

func main(){
  fc := flags.New()
  fc.NewStringFlag("password", "p", "flag for password")  //name, short_name and usage of the string flag
  fc.Parse(os.Args...)  //parse the OS arguments
  println("Flag 'password' is set: ", fc.IsSet("s"))
  println("Flag 'password' value: ", fc.String("s"))
}
```
Running the above code
```
$ main -password abc
Flag 'password' is set: true
Flag 'password' value: abc
```

# Available Flag Constructor
Flags: String, Int, float64, Bool, String Slice
```Go
NewStringFlag(name string, short_name string, usage string)
NewStringFlagWithDefault(name string, short_name string, usage string, value string)
NewIntFlag(name string, short_name string, usage string)
NewIntFlagWithDefault(name string, short_name string, usage string, value int)
NewFloat64Flag(name string, short_name string, usage string)
NewFloat64FlagWithDefault(name string, short_name string, usage string, value float64)
NewStringSliceFlag(name string, short_name string, usage string) //this flag can be supplied more than 1 time
NewStringSliceFlagWithDefault(name string, short_name string, usage string, value []string)
NewBoolFlag(name string, short_name string, usage string)
```

# Functions for flags/args reading
```Go
IsSet(flag_name string)bool
String(flag_name string)string
Int(flag_name string)int
Float64(flag_name string)float64
Bool(flag_name string)bool
StringSlice(flag_name string)[]string  
Args()[]string
```

# Parsing flags and arguments
```Go
Parse(args ...string)error  //returns error for any non-defined flags & invalid value for Int, Float64 and Bool flag.
```
Sample Code
```Go
fc := flags.New()
fc.NewIntFlag("i", "", "Int flag name i")  //set up a Int flag '-i'
fc.NewBoolFlag("verbose", "v", "Bool flag name verbose")  //set up a bool flag '-verbose'
err := fc.Parse(os.Args...) //Parse() returns any error it finds during parsing
If err != nil {
  fmt.Println("Parsing error:", err)
}
fmt.Println("Args:", fc.Args())  //Args() returns an array of all the arguments
fmt.Println("Verbose:", fc.Bool("verbose"))
fmt.Println("i:", fc.Int("i"))
```
Running above
```bash
$ app arg_1 -i 100 arg_2 -verbose  # run the code
Args: [arg_1 arg_2]
Verbose: true
i: 100
```

# Special function
```Go
SkipFlagParsing(bool)  //if set to true, all flags become arguments
ShowUsage(leadingSpace int)string  //string containing all the flags and their usage text
```
