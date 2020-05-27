package matchers_test

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

type dummyStruct struct{}
type otherStruct struct{}

var _ = Describe("HaveTypeOf", func() {
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
			expected = dummyStruct{}
		})

		JustBeforeEach(func() {
			matcher = HaveTypeOf(expected)

			success, executeErr = matcher.Match(actual)
		})

		When("actual does match expected", func() {
			BeforeEach(func() {
				actual = dummyStruct{}
			})

			It("returns true and no error", func() {
				Expect(success).To(BeTrue())
				Expect(executeErr).To(BeNil())
			})
		})

		When("actual is the wrong primitive type", func() {
			BeforeEach(func() {
				actual = "a string"
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(BeNil())
			})
		})

		When("actual is the wrong struct type", func() {
			BeforeEach(func() {
				actual = otherStruct{}
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(BeNil())
			})
		})

		When("expected is a pointer, but actual is not", func() {
			BeforeEach(func() {
				expected = &dummyStruct{}
				actual = dummyStruct{}
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(BeNil())
			})
		})
	})

	Describe("FailureMessage", func() {
		It("shows expected and actual", func() {
			expected := dummyStruct{}
			actual := otherStruct{}

			matcher = HaveTypeOf(expected)
			matcher.Match(actual)
			Expect(matcher.FailureMessage(actual)).To(MatchRegexp("(?s).*to have type:.*dummyStruct.*but it had type:.*otherStruct"))
		})
	})

	Describe("NegatedFailureMessage", func() {
		It("shows expected and actual", func() {
			expected := dummyStruct{}
			actual := dummyStruct{}

			matcher = HaveTypeOf(expected)
			matcher.Match(actual)
			Expect(matcher.NegatedFailureMessage(actual)).To(MatchRegexp("(?s).*not to have type:.*dummyStruct.*but it had type:.*dummyStruct"))
		})
	})
})
