package patch

import (
	"fmt"
)

type ArrayIndex struct {
	Index     int
	Modifiers []Modifier
	Array     []interface{}
}

func (i ArrayIndex) Concrete() (int, error) {
	result := i.Index

	for _, modifier := range i.Modifiers {
		switch modifier.(type) {
		case PrevModifier:
			result -= 1
		case NextModifier:
			result += 1
		default:
			return 0, fmt.Errorf("Expected to find one of the following modifiers: 'prev', 'next', but found modifier '%T'", modifier)
		}
	}

	if result >= len(i.Array) || (-result)-1 >= len(i.Array) {
		return 0, OpMissingIndexErr{result, i.Array}
	}

	if result < 0 {
		result = len(i.Array) + result
	}

	return result, nil
}
