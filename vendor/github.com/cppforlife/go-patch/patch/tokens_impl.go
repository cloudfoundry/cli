package patch

var _ Token = RootToken{}
var _ Token = IndexToken{}
var _ Token = AfterLastIndexToken{}
var _ Token = MatchingIndexToken{}
var _ Token = KeyToken{}

func (RootToken) _token()           {}
func (IndexToken) _token()          {}
func (AfterLastIndexToken) _token() {}
func (MatchingIndexToken) _token()  {}
func (KeyToken) _token()            {}

var _ Modifier = PrevModifier{}
var _ Modifier = NextModifier{}
var _ Modifier = BeforeModifier{}
var _ Modifier = AfterModifier{}

func (PrevModifier) _modifier()   {}
func (NextModifier) _modifier()   {}
func (BeforeModifier) _modifier() {}
func (AfterModifier) _modifier()  {}
