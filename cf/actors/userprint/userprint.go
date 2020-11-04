package userprint

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . UserPrinter

type UserPrinter interface {
	PrintUsers(guid string, username string)
}
