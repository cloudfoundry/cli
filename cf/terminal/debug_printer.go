package terminal

import (
	"time"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/trace"
)

type DebugPrinter struct {
	Logger trace.Printer
}

func (p DebugPrinter) Print(title, dump string) {
	p.Logger.Printf("\n%s [%s]\n%s\n", HeaderColor(T(title)), time.Now().Format(time.RFC3339), trace.Sanitize(dump))
}
