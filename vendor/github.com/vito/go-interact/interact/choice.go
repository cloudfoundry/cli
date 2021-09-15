package interact

// Choice is used to allow the user to select a value of an arbitrary type.
// Its Display value will be shown in a listing during Resolve, and if its
// entry in the list is chosen, the Value will be used to populate the
// destination.
type Choice struct {
	Display string
	Value   interface{}
}
