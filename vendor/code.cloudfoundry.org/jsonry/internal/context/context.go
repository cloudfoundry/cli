package context

import (
	"fmt"
	"reflect"
)

type sort uint

const (
	field sort = iota
	index
	key
)

type Context []segment

func (ctx Context) WithField(n string, t reflect.Type) Context {
	return append(ctx, segment{sort: field, name: n, typ: t})
}

func (ctx Context) WithIndex(i int, t reflect.Type) Context {
	return append(ctx, segment{sort: index, index: i, typ: t})
}

func (ctx Context) WithKey(k string, t reflect.Type) Context {
	return append(ctx, segment{sort: key, name: k, typ: t})
}

func (ctx Context) String() string {
	switch len(ctx) {
	case 0:
		return "root path"
	case 1:
		return ctx.leaf().String()
	default:
		return fmt.Sprintf("%s path %s", ctx.leaf(), ctx.path())
	}
}

func (ctx Context) leaf() segment {
	return ctx[len(ctx)-1]
}

func (ctx Context) path() string {
	var path string
	for _, s := range ctx {
		switch s.sort {
		case index:
			path = fmt.Sprintf("%s[%d]", path, s.index)
		case field:
			if len(path) > 0 {
				path = path + "."
			}
			path = path + s.name
		case key:
			path = fmt.Sprintf(`%s["%s"]`, path, s.name)
		}
	}

	return path
}
