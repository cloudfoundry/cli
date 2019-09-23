package matchers

import (
	"errors"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"reflect"
)

type ContainElementsInOrderMatcher struct {
	Elements interface{}
}

func ContainElementsInOrder(elements ...interface{}) gomega.OmegaMatcher {
	return &ContainElementsInOrderMatcher{
		Elements: elements,
	}
}

func (matcher *ContainElementsInOrderMatcher) Match(actual interface{}) (success bool, err error) {
	if !isArrayOrSlice(actual) || !isArrayOrSlice(matcher.Elements) {
		return false, errors.New("expected an array")
	}

	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(matcher.Elements)

	index := 0

OUTER:
	for i := 0; i < expectedValue.Len(); i++ {
		for ; index < actualValue.Len(); index++ {
			if reflect.DeepEqual(expectedValue.Index(i).Interface(), actualValue.Index(index).Interface()) {
				index = index + 1
				continue OUTER
			}
		}

		return false, nil
	}

	return true, nil
}

func (matcher *ContainElementsInOrderMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to contain elements in order", matcher.Elements)
}

func (matcher *ContainElementsInOrderMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to contain elements in order", matcher.Elements)
}

func isArrayOrSlice(a interface{}) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}
