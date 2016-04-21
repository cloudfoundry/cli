package userprint

//go:generate counterfeiter . UserPrinter

type UserPrinter interface {
	PrintUsers(guid string, username string)
}
