package matchers_test

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"errors"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var _ = Describe("MatchPath", func() {
	var (
		matcher  gomega.OmegaMatcher
		actual   interface{}
		expected string
	)

	BeforeEach(func() {
		expected = "/path/to/expected"
	})

	Describe("Match", func() {
		var (
			success    bool
			executeErr error
		)

		JustBeforeEach(func() {
			matcher = MatchPath(expected)

			success, executeErr = matcher.Match(actual)
		})

		When("Actual is not a string", func() {
			BeforeEach(func() {
				actual = true
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(MatchError(errors.New("MatchPath: Actual must be a string, got bool")))
			})
		})

		When("Actual does not match expected", func() {
			BeforeEach(func() {
				actual = "/path/to/something/else"
			})

			It("returns an error", func() {
				Expect(success).To(BeFalse())
				Expect(executeErr).To(BeNil())
			})
		})

		When("Actual does match expected", func() {
			When("Not Windows", func() {
				BeforeEach(func() {
					actual = "/PATH/to/expected"
				})

				It("matches successfully", func() {
					Expect(success).To(BeTrue())
					Expect(executeErr).To(BeNil())
				})
			})

			When("Windows", func() {
				BeforeEach(func() {
					actual = `\\PATH\\to\\expected`
					expected = `\\PATH\\to\\expected`
				})
				It("matches successfully", func() {
					Expect(success).To(BeTrue())
					Expect(executeErr).To(BeNil())
				})
			})
		})
	})

	Describe("FailureMessage", func() {
		It("shows expected and actual", func() {
			actual = "actual"
			expected = "expected"
			matcher = MatchPath(expected)
			matcher.Match(actual)
			Expect(matcher.FailureMessage("does not matter")).To(MatchRegexp("(?s)expected.*to match actual.*actual"))
		})
	})

	Describe("NegatedFailureMessage", func() {
		It("shows expected and actual", func() {
			actual = "actual"
			expected = "expected"
			matcher = MatchPath(expected)
			matcher.Match(actual)
			Expect(matcher.NegatedFailureMessage("does not matter")).To(MatchRegexp("(?s)expected.*not to match actual.*actual"))
		})
	})
})
