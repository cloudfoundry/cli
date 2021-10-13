package patch

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OpDefinition struct is useful for JSON and YAML unmarshaling
type OpDefinition struct {
	Type  string       `json:",omitempty"`
	Path  *string      `json:",omitempty"`
	Value *interface{} `json:",omitempty"`

	Error *string `json:",omitempty"`
}

type parser struct{}

func NewOpsFromDefinitions(opDefs []OpDefinition) (Ops, error) {
	var ops []Op
	var p parser

	for i, opDef := range opDefs {
		var op Op
		var err error

		opFmt := p.fmtOpDef(opDef)

		switch opDef.Type {
		case "replace":
			op, err = p.newReplaceOp(opDef)
			if err != nil {
				return nil, fmt.Errorf("Replace operation [%d]: %s within\n%s", i, err, opFmt)
			}

		case "remove":
			op, err = p.newRemoveOp(opDef)
			if err != nil {
				return nil, fmt.Errorf("Remove operation [%d]: %s within\n%s", i, err, opFmt)
			}

		default:
			return nil, fmt.Errorf("Unknown operation [%d] with type '%s' within\n%s", i, opDef.Type, opFmt)
		}

		if opDef.Error != nil {
			op = DescriptiveOp{Op: op, ErrorMsg: *opDef.Error}
		}

		ops = append(ops, op)
	}

	return Ops(ops), nil
}

func (parser) newReplaceOp(opDef OpDefinition) (ReplaceOp, error) {
	if opDef.Path == nil {
		return ReplaceOp{}, fmt.Errorf("Missing path")
	}

	if opDef.Value == nil {
		return ReplaceOp{}, fmt.Errorf("Missing value")
	}

	ptr, err := NewPointerFromString(*opDef.Path)
	if err != nil {
		return ReplaceOp{}, fmt.Errorf("Invalid path: %s", err)
	}

	return ReplaceOp{Path: ptr, Value: *opDef.Value}, nil
}

func (parser) newRemoveOp(opDef OpDefinition) (RemoveOp, error) {
	if opDef.Path == nil {
		return RemoveOp{}, fmt.Errorf("Missing path")
	}

	if opDef.Value != nil {
		return RemoveOp{}, fmt.Errorf("Cannot specify value")
	}

	ptr, err := NewPointerFromString(*opDef.Path)
	if err != nil {
		return RemoveOp{}, fmt.Errorf("Invalid path: %s", err)
	}

	return RemoveOp{Path: ptr}, nil
}

func (parser) fmtOpDef(opDef OpDefinition) string {
	var (
		redactedVal interface{} = "<redacted>"
		htmlDecoder             = strings.NewReplacer("\\u003c", "<", "\\u003e", ">")
	)

	if opDef.Value != nil {
		// can't JSON serialize generic interface{} anyway
		opDef.Value = &redactedVal
	}

	bytes, err := json.MarshalIndent(opDef, "", "  ")
	if err != nil {
		return "<unknown>"
	}

	return htmlDecoder.Replace(string(bytes))
}
