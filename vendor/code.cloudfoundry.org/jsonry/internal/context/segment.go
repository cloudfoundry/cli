package context

import (
	"fmt"
	"reflect"
)

type segment struct {
	sort  sort
	name  string
	index int
	typ   reflect.Type
}

func (s segment) String() string {
	switch s.sort {
	case field:
		return fmt.Sprintf(`field "%s" (type "%s")`, s.name, s.typ)
	case index:
		return fmt.Sprintf(`index %d (type "%s")`, s.index, s.typ)
	default:
		return fmt.Sprintf(`key "%s" (type "%s")`, s.name, s.typ)
	}
}
