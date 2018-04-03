package fakes

import (
	"fmt"
	"github.com/onsi/gomega/types"
)

func BeASubstageOf(expected interface{}) types.GomegaMatcher {
	return &beASubstageOfMatcher{
		parent: expected,
	}
}

type beASubstageOfMatcher struct {
	parent interface{}
}

func (matcher *beASubstageOfMatcher) Match(child interface{}) (success bool, err error) {
	substage, ok := child.(*FakeStage)
	if !ok {
		return false, fmt.Errorf("BeASubstageOf matcher expects an *fakeui.FakeStage")
	}

	parentStage, ok := matcher.parent.(*FakeStage)
	if !ok {
		return false, fmt.Errorf("BeASubstageOf matcher expects an *fakeui.FakeStage for expected value")
	}

	for _, currentSubstage := range parentStage.SubStages {
		if currentSubstage == substage {
			return true, nil
		}
	}

	return false, nil
}

func (matcher *beASubstageOfMatcher) FailureMessage(child interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be a substage of\n\t%#v", child, matcher.parent)
}

func (matcher *beASubstageOfMatcher) NegatedFailureMessage(child interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not be a substage of\n\t%#v", child, matcher.parent)
}
