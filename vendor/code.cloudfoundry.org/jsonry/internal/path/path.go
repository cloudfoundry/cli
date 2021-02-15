package path

import (
	"reflect"
	"strings"
)

const (
	omitEmptyToken  string = ",omitempty"
	omitAlwaysToken string = "-"
)

type Segment struct {
	Name string
	List bool
}

type Path struct {
	segments   []Segment
	OmitEmpty  bool
	OmitAlways bool
}

func (p Path) Len() int {
	return len(p.segments)
}

func (p Path) Pull() (Segment, Path) {
	return p.segments[0], Path{
		segments:  p.segments[1:],
		OmitEmpty: p.OmitEmpty,
	}
}

func (p Path) String() string {
	var parts []string
	for _, s := range p.segments {
		name := s.Name
		if s.List {
			name = name + "[]"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ".")
}

func ComputePath(field reflect.StructField) Path {
	var segments []Segment
	name := field.Name
	omitempty := false
	omitalways := false

	if tag := field.Tag.Get("json"); tag != "" {
		name, omitempty, omitalways = parseTag(tag, field.Name)
	} else if tag := field.Tag.Get("jsonry"); tag != "" {
		name, omitempty, omitalways = parseTag(tag, field.Name)
		segments = parseSegments(name)
	}

	if len(segments) == 0 {
		segments = append(segments, Segment{
			Name: name,
			List: false,
		})
	}

	return Path{
		OmitEmpty:  omitempty,
		OmitAlways: omitalways,
		segments:   segments,
	}
}

func parseTag(tag, defaultName string) (name string, omitempty, omitalways bool) {
	if tag == omitAlwaysToken {
		return defaultName, false, true
	}

	commaIndex := strings.IndexByte(tag, ',')
	switch commaIndex {
	case -1:
		name = tag
	case 0:
		name = defaultName
		omitempty = tag == omitEmptyToken
	default:
		name = tag[0:commaIndex]
		omitempty = tag[commaIndex:] == omitEmptyToken
	}

	return
}

func parseSegments(name string) (s []Segment) {
	add := func(elem string) {
		if strings.HasSuffix(elem, "[]") {
			s = append(s, Segment{Name: elem[:len(elem)-2], List: true})
		} else {
			s = append(s, Segment{Name: elem})
		}
	}

	start := 0
	delim := strings.IndexByte(name, '.')
	for delim >= 0 {
		add(name[start : start+delim])
		start = start + delim + 1
		delim = strings.IndexByte(name[start:], '.')
	}

	add(name[start:])

	return
}
