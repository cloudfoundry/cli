package fakes

import (
	"fmt"
)

type FakeGenerator struct {
	GeneratedUUID string
	NextUUID      int
	GenerateError error
}

func NewFakeGenerator() *FakeGenerator {
	return &FakeGenerator{}
}

func (gen *FakeGenerator) Generate() (uuid string, err error) {
	if gen.GeneratedUUID == "" && gen.GenerateError == nil {
		uuidString := fmt.Sprintf("fake-uuid-%d", gen.NextUUID)
		gen.NextUUID++
		return uuidString, nil
	}
	return gen.GeneratedUUID, gen.GenerateError
}
