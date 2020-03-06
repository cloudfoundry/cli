package jsonry

import (
	"reflect"
	"strings"
)

type pathElement struct {
	name string
	list bool
}

type jsonryPath struct {
	elements  []pathElement
	omitempty bool
}

func (p jsonryPath) chop() jsonryPath {
	return jsonryPath{
		elements:  p.elements[1:],
		omitempty: p.omitempty,
	}
}

func (p jsonryPath) len() int {
	return len(p.elements)
}

func actualKind(t reflect.Type) reflect.Kind {
	if k := t.Kind(); k != reflect.Ptr {
		return k
	}

	return t.Elem().Kind()
}

func computePath(field reflect.StructField) jsonryPath {
	if tag := field.Tag.Get("jsonry"); tag != "" {
		parts := strings.Split(tag, ",")

		jsonry := parts[0]
		if len(jsonry) == 0 {
			jsonry = strings.ToLower(field.Name)
		}

		var elements []pathElement
		for _, elem := range strings.Split(jsonry, ".") {
			elements = append(elements, pathElement{
				name: strings.TrimRight(elem, "[]"),
				list: strings.HasSuffix(elem, "[]"),
			})
		}
		return jsonryPath{
			elements:  elements,
			omitempty: len(parts) > 1 && parts[1] == "omitempty",
		}
	}

	if tag := field.Tag.Get("json"); tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return jsonryPath{elements: []pathElement{{name: parts[0]}}}
		}
	}

	return jsonryPath{elements: []pathElement{{name: strings.ToLower(field.Name)}}}
}
