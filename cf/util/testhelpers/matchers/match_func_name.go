package matchers

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/onsi/gomega"
)

type FuncsByNameMatcher struct {
	expected      []interface{}
	actualNames   []string
	expectedNames []string
}

func MatchFuncsByName(funcs ...interface{}) gomega.OmegaMatcher {
	return &FuncsByNameMatcher{expected: funcs}
}

func (matcher *FuncsByNameMatcher) Match(actual interface{}) (success bool, err error) {
	for _, fn := range matcher.expected {
		// just to be safe
		if v := reflect.ValueOf(fn); v.Kind() == reflect.Func {
			name := runtime.FuncForPC(v.Pointer()).Name()
			matcher.expectedNames = append(matcher.expectedNames, name)
		} else {
			return false, fmt.Errorf("MatchChangeAppFuncsByName: Expected must be a slice of functions, got %s", v.Type().Name())
		}
	}

	if t := reflect.TypeOf(actual); t.Kind() != reflect.Slice {
		return false, fmt.Errorf("MatchChangeAppFuncsByName: Actual must be a slice of functions, got %s", t.Name())
	}

	arr := reflect.ValueOf(actual)

	for i := 0; i < arr.Len(); i++ {
		elem := arr.Index(i)
		if elem.Kind() != reflect.Func {
			return false, fmt.Errorf("MatchChangeAppFuncsByName: Actual must be a slice of functions, got %s", elem.Type().Name())
		}
		matcher.actualNames = append(matcher.actualNames, runtime.FuncForPC(elem.Pointer()).Name())
	}

	return reflect.DeepEqual(matcher.actualNames, matcher.expectedNames), nil
}

func (matcher *FuncsByNameMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n to match actual:\n%v\n", matcher.expectedNames, matcher.actualNames)
}

func (matcher *FuncsByNameMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n %v\n not to match actual:\n%v\n", matcher.expectedNames, matcher.actualNames)
}
