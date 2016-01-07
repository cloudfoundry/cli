package flags

import "strings"

func (c *flagContext) ShowUsage(leadingSpace int) string {
	output := ""
	names := make([][]string, len(c.cmdFlags))

	var l, i int
	for _, f := range c.cmdFlags {
		var line string
		if f.GetName() != "" {
			line += "--" + f.GetName()
			if f.GetShortName() != "" {
				line += ", "
			}
		}

		if f.GetShortName() != "" {
			line += "-" + f.GetShortName()
		}

		if len(line) > l {
			l = len(line)
		}

		names[i] = []string{
			strings.Repeat(" ", leadingSpace) + line,
			f.String(),
		}

		i++
	}

	for _, atts := range names {
		flagList := atts[0]
		usage := atts[1]
		line := flagList

		if usage != "" {
			line += strings.Repeat(" ", 6+(l-len(flagList))) + usage
		}

		output += line + "\n"
	}

	return output
}
