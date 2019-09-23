package matchers

import (
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"reflect"
)

type ContainElementTimesMatcher struct {
	Element       interface{}
	ExpectedCount int
	ActualCount   int
}

func ContainElementTimes(element interface{}, count int) gomega.OmegaMatcher {
	return &ContainElementTimesMatcher{
		Element:       element,
		ExpectedCount: count,
	}
}

func (matcher *ContainElementTimesMatcher) Match(actual interface{}) (success bool, err error) {
	if !isArrayOrSlice(actual) {
		return false, errors.New("expected an array")
	}

	actualValue := reflect.ValueOf(actual)

	for i := 0; i < actualValue.Len(); i++ {
		if reflect.DeepEqual(actualValue.Index(i).Interface(), matcher.Element) {
			matcher.ActualCount++
		}
	}

	return matcher.ActualCount == matcher.ExpectedCount, nil
}

func (matcher *ContainElementTimesMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf(
			"to contain element %v %d times, but found it %d times",
			matcher.Element,
			matcher.ExpectedCount,
			matcher.ActualCount,
		),
	)
}

func (matcher *ContainElementTimesMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf(
			"not to contain element %v %d times, but it did",
			matcher.Element,
			matcher.ExpectedCount,
		),
	)
}
