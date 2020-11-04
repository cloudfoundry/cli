package pushaction

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RandomWordGenerator

// RandomWordGenerator returns random words.
type RandomWordGenerator interface {
	RandomAdjective() string
	RandomNoun() string
	RandomTwoLetters() string
}
