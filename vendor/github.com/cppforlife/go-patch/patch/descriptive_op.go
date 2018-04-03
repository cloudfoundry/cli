package patch

import (
	"fmt"
)

type DescriptiveOp struct {
	Op       Op
	ErrorMsg string
}

func (op DescriptiveOp) Apply(doc interface{}) (interface{}, error) {
	doc, err := op.Op.Apply(doc)
	if err != nil {
		return nil, fmt.Errorf("Error '%s': %s", op.ErrorMsg, err.Error())
	}
	return doc, nil
}
