package patch

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type ReplaceOp struct {
	Path  Pointer
	Value interface{} // will be cloned using yaml library
}

func (op ReplaceOp) Apply(doc interface{}) (interface{}, error) {
	// Ensure that value is not modified by future operations
	clonedValue, err := op.cloneValue(op.Value)
	if err != nil {
		return nil, fmt.Errorf("ReplaceOp cloning value: %s", err)
	}

	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return clonedValue, nil
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

			if isLast {
				idx, err := ArrayInsertion{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
				if err != nil {
					return nil, err
				}

				prevUpdate(idx.Update(typedObj, clonedValue))
			} else {
				idx, err := ArrayIndex{Index: typedToken.Index, Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
				if err != nil {
					return nil, err
				}

				obj = typedObj[idx]
				prevUpdate = func(newObj interface{}) { typedObj[idx] = newObj }
			}

		case AfterLastIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			if isLast {
				prevUpdate(append(typedObj, clonedValue))
			} else {
				return nil, fmt.Errorf("Expected after last index token to be last in path '%s'", op.Path)
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
				if isLast {
					prevUpdate(append(typedObj, clonedValue))
				} else {
					obj = map[interface{}]interface{}{typedToken.Key: typedToken.Value}
					prevUpdate(append(typedObj, obj))
					// no need to change prevUpdate since matching item can only be a map
				}
			} else {
				if len(idxs) != 1 {
					return nil, opMultipleMatchingIndexErr{NewPointer(tokens[:i+2]), idxs}
				}

				if isLast {
					idx, err := ArrayInsertion{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
					if err != nil {
						return nil, err
					}

					prevUpdate(idx.Update(typedObj, clonedValue))
				} else {
					idx, err := ArrayIndex{Index: idxs[0], Modifiers: typedToken.Modifiers, Array: typedObj}.Concrete()
					if err != nil {
						return nil, err
					}

					obj = typedObj[idx]
					// no need to change prevUpdate since matching item can only be a map
				}
			}

		case KeyToken:
			typedObj, ok := obj.(map[interface{}]interface{})
			if !ok {
				return nil, newOpMapMismatchTypeErr(tokens[:i+2], obj)
			}

			var found bool

			obj, found = typedObj[typedToken.Key]
			if !found && !typedToken.Optional {
				return nil, opMissingMapKeyErr{typedToken.Key, NewPointer(tokens[:i+2]), typedObj}
			}

			if isLast {
				typedObj[typedToken.Key] = clonedValue
			} else {
				prevUpdate = func(newObj interface{}) { typedObj[typedToken.Key] = newObj }

				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case AfterLastIndexToken:
						obj = []interface{}{}
					case MatchingIndexToken:
						obj = []interface{}{}
					case KeyToken:
						obj = map[interface{}]interface{}{}
					default:
						errMsg := "Expected to find key, matching index or after last index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}

					typedObj[typedToken.Key] = obj
				}
			}

		default:
			return nil, opUnexpectedTokenErr{token, NewPointer(tokens[:i+2])}
		}
	}

	return doc, nil
}

func (ReplaceOp) cloneValue(in interface{}) (out interface{}, err error) {
	defer func() {
		if recoverVal := recover(); recoverVal != nil {
			err = fmt.Errorf("Recovered: %s", recoverVal)
		}
	}()

	bytes, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
