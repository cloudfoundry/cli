package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"io"
	"os"
	"strings"
)

var prettifier *gosteno.JsonPrettifier

var ignoreParseError = false
var helptext = `Usage: gosteno-prettify [OPTS] [FILE(s)]

Parses json formatted log lines from FILE(s), or stdin,
and displays a more human friendly version of each line to stdout.

Examples :

    gosteno-prettify f - g
        Prettify f's contents, then standard input, then g's contents.

    gosteno-prettify
        Prettify contents of stdin.

Options:

    -h
        Display help
    -a
        Omit location and data in order to provide well-aligned logs
    -s
        Donot complain about errors in parsing logs
`

func prettifyFromIO(src io.Reader) {
	buf := bufio.NewReader(src)

	lineno := 0
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		lineno++

		i := strings.Index(line, "{")
		var prefix string
		if i > 0 {
			prefix = line[:i]
			line = line[i:]
		}
		r, e := prettifier.DecodeJsonLogEntry(line)
		if e != nil {
			if !ignoreParseError {
				fmt.Fprintf(os.Stderr, "steno-prettify: Malformed json at line %d : %s", lineno, line)
			}
			continue
		}
		s, _ := prettifier.EncodeRecord(r)
		fmt.Println(fmt.Sprintf("%s%s", prefix, string(s)))
	}
}

func prettifyFromFile(logFile string) error {
	f, e := os.Open(logFile)
	if e != nil {
		return e
	}
	defer f.Close()

	prettifyFromIO(f)

	return nil
}

func main() {
	excludedFields := gosteno.EXCLUDE_NONE

	h := flag.Bool("h", false, "Show help")
	a := flag.Bool("a", false, "Omit location and data in order to provide well-aligned logs")
	s := flag.Bool("s", false, "Ignore errors in parsing logs")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, helptext)
	}

	flag.Parse()

	if *h {
		flag.Usage()
		return
	}
	if *a {
		excludedFields |= gosteno.EXCLUDE_FILE | gosteno.EXCLUDE_LINE | gosteno.EXCLUDE_METHOD | gosteno.EXCLUDE_DATA
	}
	if *s {
		ignoreParseError = true
	}

	prettifier = gosteno.NewJsonPrettifier(excludedFields)

	args := flag.Args()
	if len(args) > 0 {
		for _, f := range args {
			if f == "-" {
				prettifyFromIO(os.Stdin)
			} else {
				e := prettifyFromFile(f)
				if e != nil {
					fmt.Fprintln(os.Stderr, e)
				}
			}
		}
	} else {
		prettifyFromIO(os.Stdin)
	}
}
