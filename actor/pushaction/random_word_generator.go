package pushaction

//go:generate counterfeiter . RandomWordGenerator

// RandomWordGenerator returns random words.
type RandomWordGenerator interface {
	RandomAdjective() string
	RandomNoun() string
}
