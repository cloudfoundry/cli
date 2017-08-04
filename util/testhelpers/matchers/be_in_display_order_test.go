package matchers_test

import (
	"strings"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BeInDisplayOrder()", func() {
	var (
		matcher OmegaMatcher
		actual  []string
	)

	actual = []string{
		"1st line 1",
		"2nd line 2",
		"3rd line 3",
		"4th line 4",
	}

	It("asserts actual is in same display order with expected", func() {
		matcher = BeInDisplayOrder(
			[]string{"1st"},
			[]string{"2nd"},
			[]string{"3rd"},
			[]string{"4th"},
		)

		success, err := matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(true))

		matcher = BeInDisplayOrder(
			[]string{"1st"},
			[]string{"3rd"},
			[]string{"2nd"},
			[]string{"4th"},
		)

		success, err = matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(false))
		msg := matcher.FailureMessage([]string{})
		Expect(strings.Contains(msg, "2nd")).To(Equal(true))
	})

	It("asserts actual contains the expected string", func() {
		matcher = BeInDisplayOrder(
			[]string{"Not in the actual"},
		)

		success, err := matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(false))

		msg := matcher.FailureMessage([]string{})
		Expect(strings.Contains(msg, "Not in the actual")).To(Equal(true))
	})

	It("asserts actual contains 2 substrings in the same display line ", func() {
		matcher = BeInDisplayOrder(
			[]string{"1st", "line"},
			[]string{"4th", "line"},
		)

		success, err := matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(true))

		matcher = BeInDisplayOrder(
			[]string{"1st", "line 2"},
		)

		success, err = matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(false))

		msg := matcher.FailureMessage([]string{})
		Expect(strings.Contains(msg, "line 2")).To(Equal(true))

		matcher = BeInDisplayOrder(
			[]string{"1st"},
			[]string{"line 1"},
		)

		success, err = matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(false))

		msg = matcher.FailureMessage([]string{})
		Expect(strings.Contains(msg, "line 1")).To(Equal(true))
	})

	It("asserts actual contains 2 substrings displaying in order on a single line ", func() {
		matcher = BeInDisplayOrder(
			[]string{"1st", "line 1"},
		)

		success, err := matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(true))

		matcher = BeInDisplayOrder(
			[]string{"line 1", "1st"},
		)

		success, err = matcher.Match(actual)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(Equal(false))

		msg := matcher.FailureMessage([]string{})
		Expect(strings.Contains(msg, "1st")).To(Equal(true))
	})
})
