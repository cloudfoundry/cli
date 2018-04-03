package patch

import (
	"fmt"
	"sort"
	"strings"
)

type opMismatchTypeErr struct {
	type_ string
	path  Pointer
	obj   interface{}
}

func newOpArrayMismatchTypeErr(tokens []Token, obj interface{}) opMismatchTypeErr {
	return opMismatchTypeErr{"an array", NewPointer(tokens), obj}
}

func newOpMapMismatchTypeErr(tokens []Token, obj interface{}) opMismatchTypeErr {
	return opMismatchTypeErr{"a map", NewPointer(tokens), obj}
}

func (e opMismatchTypeErr) Error() string {
	errMsg := "Expected to find %s at path '%s' but found '%T'"
	return fmt.Sprintf(errMsg, e.type_, e.path, e.obj)
}

type opMissingMapKeyErr struct {
	key  string
	path Pointer
	obj  map[interface{}]interface{}
}

func (e opMissingMapKeyErr) Error() string {
	errMsg := "Expected to find a map key '%s' for path '%s' (%s)"
	return fmt.Sprintf(errMsg, e.key, e.path, e.siblingKeysErrStr())
}

func (e opMissingMapKeyErr) siblingKeysErrStr() string {
	if len(e.obj) == 0 {
		return "found no other map keys"
	}
	var keys []string
	for key, _ := range e.obj {
		if keyStr, ok := key.(string); ok {
			keys = append(keys, keyStr)
		}
	}
	sort.Sort(sort.StringSlice(keys))
	return "found map keys: '" + strings.Join(keys, "', '") + "'"
}

type OpMissingIndexErr struct {
	Idx int
	Obj []interface{}
}

func (e OpMissingIndexErr) Error() string {
	return fmt.Sprintf("Expected to find array index '%d' but found array of length '%d'", e.Idx, len(e.Obj))
}

type opMultipleMatchingIndexErr struct {
	path Pointer
	idxs []int
}

func (e opMultipleMatchingIndexErr) Error() string {
	return fmt.Sprintf("Expected to find exactly one matching array item for path '%s' but found %d", e.path, len(e.idxs))
}

type opUnexpectedTokenErr struct {
	token Token
	path  Pointer
}

func (e opUnexpectedTokenErr) Error() string {
	return fmt.Sprintf("Expected to not find token '%T' at path '%s'", e.token, e.path)
}
