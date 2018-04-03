package index

import (
	"errors"
)

var (
	ErrNotFound = errors.New("Record is not found")
)

type Index interface {
	Find(interface{}, interface{}) error
	Save(interface{}, interface{}) error
}
