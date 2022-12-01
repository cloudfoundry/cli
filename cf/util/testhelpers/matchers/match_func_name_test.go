package matchers_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func dummyFunc()         {}
func dummyNotMatchFunc() {}

var _ = Describe("MatchChangeAppFuncsByName", func() {
	var (
		matcher  gomega.OmegaMatcher
		actual   interface{}
		expected interface{}
	)

	Describe("Match", func() {
		var (
			success    bool
			executeErr error
		)

		BeforeEach(func() {
			expected = dummyFunc
		})

		JustBeforeEach(func() {
			matcher = MatchFuncsByName(expected)

			success, executeErr = matcher.Match(actual)
		})

		When("Expected is not a list of funcs", func() {
			BeforeEach(func() {
				expected = "hello"
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(MatchError(errors.New("MatchChangeAppFuncsByName: Expected must be a slice of functions, got string")))
			})
		})

		When("Actual is not a slice", func() {
			BeforeEach(func() {
				actual = true
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(MatchError(errors.New("MatchChangeAppFuncsByName: Actual must be a slice of functions, got bool")))
			})
		})

		When("Actual is not a slice of funcs", func() {
			BeforeEach(func() {
				actual = []int{5}
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(MatchError(errors.New("MatchChangeAppFuncsByName: Actual must be a slice of functions, got int")))
			})
		})

		When("Actual does not match expected", func() {
			BeforeEach(func() {
				actual = []func(){dummyNotMatchFunc}
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(BeNil())
			})
		})

		When("Actual does match expected", func() {
			BeforeEach(func() {
				actual = []func(){dummyFunc}
			})

			It("returns an error", func() {
				Expect(success).To(BeTrue())
				Expect(executeErr).To(BeNil())
			})
		})
	})

	Describe("FailureMessage", func() {
		It("shows expected and actual", func() {
			matcher = MatchFuncsByName(dummyFunc)
			matcher.Match([]func(){dummyNotMatchFunc})
			Expect(matcher.FailureMessage("does not matter")).To(MatchRegexp("(?s)dummyFunc.*to match actual.*dummyNotMatchFunc"))
		})
	})

	Describe("NegatedFailureMessage", func() {
		It("shows expected and actual", func() {
			matcher = MatchFuncsByName(dummyFunc)
			matcher.Match([]func(){dummyNotMatchFunc})
			Expect(matcher.NegatedFailureMessage("does not matter")).To(MatchRegexp("(?s)dummyFunc.*not to match actual.*dummyNotMatchFunc"))
		})
	})
})
