package patch

type Token interface {
	_token()
}

type RootToken struct{}

type IndexToken struct {
	Index     int
	Modifiers []Modifier
}

type AfterLastIndexToken struct{}

type MatchingIndexToken struct {
	Key       string
	Value     string
	Optional  bool
	Modifiers []Modifier
}

type KeyToken struct {
	Key      string
	Optional bool
}

type Modifier interface {
	_modifier()
}

type PrevModifier struct{}
type NextModifier struct{}

type BeforeModifier struct{}
type AfterModifier struct{}
