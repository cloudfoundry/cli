package multiopt

import (
	"strings"

	"github.com/pborman/getopt"
)

type MultiValue struct {
	values []string
}

func (mv *MultiValue) Set(value string, option getopt.Option) error {
	mv.values = append(mv.values, value)
	return nil
}

func (mv *MultiValue) String() string {
	return strings.Join(mv.values, ",")
}

func (mv *MultiValue) Values() []string {
	return mv.values
}
