package patch

import (
	"fmt"
)

type ArrayInsertion struct {
	Index     int
	Modifiers []Modifier
	Array     []interface{}
}

type ArrayInsertionIndex struct {
	number int
	insert bool
}

func (i ArrayInsertion) Concrete() (ArrayInsertionIndex, error) {
	var mods []Modifier

	before := false
	after := false

	for _, modifier := range i.Modifiers {
		if before {
			return ArrayInsertionIndex{}, fmt.Errorf(
				"Expected to not find any modifiers after 'before' modifier, but found modifier '%T'", modifier)
		}
		if after {
			return ArrayInsertionIndex{}, fmt.Errorf(
				"Expected to not find any modifiers after 'after' modifier, but found modifier '%T'", modifier)
		}

		switch modifier.(type) {
		case BeforeModifier:
			before = true
		case AfterModifier:
			after = true
		default:
			mods = append(mods, modifier)
		}
	}

	idx := ArrayIndex{Index: i.Index, Modifiers: mods, Array: i.Array}

	num, err := idx.Concrete()
	if err != nil {
		return ArrayInsertionIndex{}, err
	}

	if after && num != len(i.Array) {
		num += 1
	}

	return ArrayInsertionIndex{num, before || after}, nil
}

func (i ArrayInsertionIndex) Update(array []interface{}, obj interface{}) []interface{} {
	if i.insert {
		var newAry []interface{}
		newAry = append(newAry, array[:i.number]...) // not inclusive
		newAry = append(newAry, obj)
		newAry = append(newAry, array[i.number:]...) // inclusive
		return newAry
	}

	array[i.number] = obj
	return array
}
