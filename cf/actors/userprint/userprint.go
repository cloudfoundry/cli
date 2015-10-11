package userprint

type UserPrinter interface {
	PrintUsers(guid string, username string)
}
