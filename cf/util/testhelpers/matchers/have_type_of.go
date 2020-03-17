package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega"
)

type HaveTypeOfMatcher struct {
	expected interface{}
	actual   interface{}
}

func HaveTypeOf(expected interface{}) gomega.OmegaMatcher {
	return &HaveTypeOfMatcher{expected: expected}
}

func (matcher *HaveTypeOfMatcher) Match(actual interface{}) (success bool, err error) {
	expectedType := reflect.TypeOf(matcher.expected)
	actualType := reflect.TypeOf(actual)

	return reflect.DeepEqual(actualType, expectedType), nil
}

func (matcher *HaveTypeOfMatcher) FailureMessage(actual interface{}) string {
	expectedType := getTypeName(matcher.expected)
	actualType := getTypeName(actual)

	return fmt.Sprintf("Expected:\n  %v\nto have type:\n  %v\nbut it had type:\n  %v", matcher.actual, expectedType, actualType)
}

func (matcher *HaveTypeOfMatcher) NegatedFailureMessage(actual interface{}) string {
	expectedType := getTypeName(matcher.expected)
	actualType := getTypeName(actual)

	return fmt.Sprintf("Expected:\n  %v\nnot to have type:\n  %v\nbut it had type:\n  %v", matcher.actual, expectedType, actualType)
}

func getTypeName(val interface{}) string {
	t := reflect.TypeOf(val)

	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}

	return t.Name()
}
