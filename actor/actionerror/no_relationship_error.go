package actionerror

type NoRelationshipError struct {
}

func (e NoRelationshipError) Error() string {
	return "No relationship found."
}
