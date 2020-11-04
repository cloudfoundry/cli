package trace

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Printer

type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	WritesToConsole() bool
}
