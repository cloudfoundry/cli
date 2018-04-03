package patch

import (
	"fmt"
)

type RemoveOp struct {
	Path Pointer
}

func (op RemoveOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return nil, fmt.Errorf("Cannot remove entire document")
	}

	obj := doc
	prevUpdate := func(newObj interface{}) { doc = newObj }

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2

		switch typedToken := token.(type) {
		case IndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				var newAry []interface{}
				newAry = append(newAry, typedObj[:idx]...)
				newAry = append(newAry, typedObj[idx+1:]...)
				prevUpdate(newAry)
			} else {
				obj = typedObj[idx]
				prevUpdate = func(newObj interface{}) { typedObj[idx] = newObj }
			}

		case MatchingIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			var idxs []int

			for itemIdx, item := range typedObj {
				typedItem, ok := item.(map[interface{}]interface{})
				if ok {
					if typedItem[typedToken.Key] == typedToken.Value {
						idxs = append(idxs, itemIdx)
					}
				}
			}

			if typedToken.Optional && len(idxs) == 0 {
				return doc, nil
			}

			if len(idxs) != 1 {
				return nil, opMultipleMatchingIndexErr{NewPointer(tokens[:i+2]), idxs}
			}

			idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
			if err != nil {
				return nil, err
			}

			if isLast {
				var newAry []interface{}
				newAry = append(newAry, typedObj[:idx]...)
				newAry = append(newAry, typedObj[idx+1:]...)
				prevUpdate(newAry)
			} else {
				obj = typedObj[idx]
				// no need to change prevUpdate since matching item can only be a map
			}

		case KeyToken:
			typedObj, ok := obj.(map[interface{}]interface{})
			if !ok {
				return nil, newOpMapMismatchTypeErr(tokens[:i+2], obj)
			}

			var found bool

			obj, found = typedObj[typedToken.Key]
			if !found {
				if typedToken.Optional {
					return doc, nil
				}

				return nil, opMissingMapKeyErr{typedToken.Key, NewPointer(tokens[:i+2]), typedObj}
			}

			if isLast {
				delete(typedObj, typedToken.Key)
			} else {
				prevUpdate = func(newObj interface{}) { typedObj[typedToken.Key] = newObj }
			}

		default:
			return nil, opUnexpectedTokenErr{token, NewPointer(tokens[:i+2])}
		}
	}

	return doc, nil
}
