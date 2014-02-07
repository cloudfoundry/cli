package net

import (
	"encoding/json"
	"reflect"
)

func NewPaginatedResources(exampleResource ModelResource) PaginatedResources {
	return PaginatedResources{
		Unmarshaler: sliceUnmarshaler{valueType: reflect.TypeOf(exampleResource)},
	}
}

type PaginatedResources struct {
	NextURL     string           `json:"next_url"`
	Unmarshaler sliceUnmarshaler `json:"resources"`
}

func (pag *PaginatedResources) Resources() []ModelResource {
	return pag.Unmarshaler.Contents
}

type ModelResource interface {
	ToFields() interface{}
}

type sliceUnmarshaler struct {
	valueType reflect.Type
	Contents  []ModelResource
}

func (this *sliceUnmarshaler) UnmarshalJSON(input []byte) (err error) {
	slice := reflect.New(reflect.SliceOf(this.valueType))
	err = json.Unmarshal(input, slice.Interface())
	if err != nil {
		return
	}

	value := reflect.Indirect(slice)
	contents := make([]ModelResource, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		contents = append(contents, value.Index(i).Interface().(ModelResource))
	}
	this.Contents = contents
	return
}
