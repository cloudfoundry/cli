package batcher_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util/batcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Batcher", func() {
	var (
		calls            [][]string
		warningsToReturn []ccv3.Warnings
		errorsToReturn   []error
	)

	fakeCallback := func(guids []string) (ccv3.Warnings, error) {
		index := len(calls)
		calls = append(calls, guids)

		var (
			warnings ccv3.Warnings
			err      error
		)

		if len(warningsToReturn) > index {
			warnings = warningsToReturn[index]
		}

		if len(errorsToReturn) > index {
			err = errorsToReturn[index]
		}

		return warnings, err
	}

	spreadGuids := func(start, end int) (result []string) {
		for i := start; i < end; i++ {
			result = append(result, fmt.Sprintf("fake-guid-%d", i))
		}
		return
	}

	BeforeEach(func() {
		calls = nil
		warningsToReturn = nil
		errorsToReturn = nil
	})

	It("calls the callback", func() {
		_, _ = batcher.RequestByGUID([]string{"one", "two", "three"}, fakeCallback)

		Expect(calls).To(HaveLen(1))
		Expect(calls[0]).To(Equal([]string{"one", "two", "three"}))
	})

	When("the guids list exceeds the batch size", func() {
		It("calls the callback multiple times", func() {
			_, _ = batcher.RequestByGUID(spreadGuids(0, 520), fakeCallback)

			Expect(calls).To(HaveLen(3))
			Expect(calls[0]).To(Equal(spreadGuids(0, 200)))
			Expect(calls[1]).To(Equal(spreadGuids(200, 400)))
			Expect(calls[2]).To(Equal(spreadGuids(400, 520)))
		})
	})

	When("the callback returns warnings", func() {
		BeforeEach(func() {
			warningsToReturn = []ccv3.Warnings{
				{"one", "two"},
				{"three", "four"},
				{},
				{"five"},
			}
		})

		It("returns all the accumulated warnings", func() {
			warnings, _ := batcher.RequestByGUID(spreadGuids(0, 960), fakeCallback)

			Expect(warnings).To(ConsistOf("one", "two", "three", "four", "five"))
		})
	})

	When("the callback returns an error", func() {
		BeforeEach(func() {
			warningsToReturn = []ccv3.Warnings{
				{"one", "two"},
				{"three", "four"},
				{"five"},
				{"six", "seven"},
			}

			errorsToReturn = []error{
				nil,
				nil,
				errors.New("bang"),
				nil,
			}
		})

		It("returns the error and accumulated warnings", func() {
			warnings, err := batcher.RequestByGUID(spreadGuids(0, 960), fakeCallback)

			Expect(calls).To(HaveLen(3))
			Expect(warnings).To(ConsistOf("one", "two", "three", "four", "five"))
			Expect(err).To(MatchError("bang"))
		})
	})
})
