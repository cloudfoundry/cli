package flags

import (
	"fmt"
	"sort"
	"strings"
)

func (c *flagContext) ShowUsage(leadingSpace int) string {
	displayFlags := flags{}

	for _, f := range c.cmdFlags {
		if !f.Visible() {
			continue
		}

		d := flagPresenter{
			flagSet: f,
		}

		displayFlags = append(displayFlags, d)
	}

	return displayFlags.toString(strings.Repeat(" ", leadingSpace))
}

type flagPresenter struct {
	flagSet FlagSet
}

func (p *flagPresenter) line(l int) string {
	flagList := p.flagList()
	usage := p.usage()
	spaces := strings.Repeat(" ", 6+(l-len(flagList)))

	return strings.TrimRight(fmt.Sprintf("%s%s%s", flagList, spaces, usage), " ")
}

func (p *flagPresenter) flagList() string {
	f := p.flagSet
	var parts []string

	if f.GetName() != "" {
		parts = append(parts, fmt.Sprintf("--%s", f.GetName()))
	}

	if f.GetShortName() != "" {
		parts = append(parts, fmt.Sprintf("-%s", f.GetShortName()))
	}

	return strings.Join(parts, ", ")
}

func (p *flagPresenter) usage() string {
	return p.flagSet.String()
}

func (p *flagPresenter) comparableString() string {
	if p.flagSet.GetName() != "" {
		return p.flagSet.GetName()
	}

	return p.flagSet.GetShortName()
}

type flags []flagPresenter

func (f flags) Len() int {
	return len(f)
}

func (f flags) Less(i, j int) bool {
	return (f[i].comparableString() < f[j].comparableString())
}

func (f flags) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f flags) toString(prefix string) string {
	sort.Sort(f)

	lines := make([]string, f.Len())
	maxLength := f.maxLineLength()

	for i, l := range f {
		lines[i] = fmt.Sprintf("%s%s", prefix, l.line(maxLength))
	}

	return strings.Join(lines, "\n")
}

func (f flags) maxLineLength() int {
	var l int

	for _, x := range f {

		lPrime := len(x.flagList())

		if lPrime > l {
			l = lPrime
		}
	}

	return l
}
