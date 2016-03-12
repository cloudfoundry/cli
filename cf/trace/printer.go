package trace

//go:generate counterfeiter -o fakes/fake_printer.go . Printer
type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	WritesToConsole() bool
}
