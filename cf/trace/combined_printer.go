package trace

type combinedPrinter []Printer

func CombinePrinters(printers []Printer) Printer {
	return combinedPrinter(printers)
}

func (p combinedPrinter) Print(v ...interface{}) {
	for _, printer := range p {
		printer.Print(v...)
	}
}

func (p combinedPrinter) Printf(format string, v ...interface{}) {
	for _, printer := range p {
		printer.Printf(format, v...)
	}
}

func (p combinedPrinter) Println(v ...interface{}) {
	for _, printer := range p {
		printer.Println(v...)
	}
}

func (p combinedPrinter) IsEnabled() bool {
	for _, printer := range p {
		if printer.IsEnabled() {
			return true
		}
	}
	return false
}
