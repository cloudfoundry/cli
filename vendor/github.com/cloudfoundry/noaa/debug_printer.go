package noaa

type DebugPrinter interface {
	Print(title, dump string)
}

type NullDebugPrinter struct {
}

func (NullDebugPrinter) Print(title, body string) {
}
