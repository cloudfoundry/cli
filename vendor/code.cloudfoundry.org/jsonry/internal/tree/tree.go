package tree

import (
	"reflect"

	"code.cloudfoundry.org/jsonry/internal/path"
)

type Tree map[string]interface{}

func (t Tree) Attach(p path.Path, v interface{}) Tree {
	switch p.Len() {
	case 0:
		panic("empty path")
	case 1:
		leaf, _ := p.Pull()
		t[leaf.Name] = v
	default:
		branch, stem := p.Pull()
		if branch.List {
			t[branch.Name] = spread(stem, v)
		} else {
			if _, ok := t[branch.Name].(Tree); !ok {
				t[branch.Name] = make(Tree)
			}
			t[branch.Name].(Tree).Attach(stem, v)
		}
	}

	return t
}

func (t Tree) Fetch(p path.Path) (interface{}, bool) {
	switch p.Len() {
	case 0:
		panic("empty path")
	case 1:
		leaf, _ := p.Pull()
		v, ok := t[leaf.Name]
		return v, ok
	default:
		branch, stem := p.Pull()
		v, ok := t[branch.Name]
		if !ok {
			return nil, false
		}

		switch vt := v.(type) {
		case map[string]interface{}:
			return Tree(vt).Fetch(stem)
		case []interface{}:
			return unspread(vt, stem), true
		default:
			return nil, false
		}
	}
}

func spread(p path.Path, v interface{}) []interface{} {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Array && vv.Kind() != reflect.Slice {
		v = []interface{}{v}
		vv = reflect.ValueOf(v)
	}

	var s []interface{}
	for i := 0; i < vv.Len(); i++ {
		s = append(s, make(Tree).Attach(p, vv.Index(i).Interface()))
	}
	return s
}

func unspread(v []interface{}, stem path.Path) []interface{} {
	l := make([]interface{}, 0, len(v))
	for i := range v {
		switch vt := v[i].(type) {
		case map[string]interface{}:
			if r, ok := Tree(vt).Fetch(stem); ok {
				l = append(l, r)
			} else {
				l = append(l, nil)
			}
		case []interface{}:
			l = append(l, unspread(vt, stem)...)
		default:
			l = append(l, v[i])
		}
	}

	return l
}
