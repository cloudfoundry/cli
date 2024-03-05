package railway_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/util/railway"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sequentially()", func() {
	Describe("with no errors", func() {
		It("calls all the tracks", func() {
			var (
				i = 0
				j = 0
			)

			_, _ = railway.Sequentially(
				func() (ccv3.Warnings, error) { i = 1; return ccv3.Warnings{}, nil },
				func() (ccv3.Warnings, error) { j = 2; return ccv3.Warnings{}, nil },
			)

			Expect(i).To(Equal(1))
			Expect(j).To(Equal(2))
		})

		It("calls the tracks in order", func() {
			i := 0

			_, _ = railway.Sequentially(
				func() (ccv3.Warnings, error) { i = 1; return ccv3.Warnings{}, nil },
				func() (ccv3.Warnings, error) { i = 2; return ccv3.Warnings{}, nil },
			)

			Expect(i).To(Equal(2))
		})
	})

	Describe("with errors", func() {
		It("returns the first error", func() {
			_, err := railway.Sequentially(
				func() (ccv3.Warnings, error) { return ccv3.Warnings{}, errors.New("error 1") },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{}, errors.New("error 2") },
			)

			Expect(err).To(Equal(errors.New("error 1")))
		})

		It("runs all tracks before the error", func() {
			i := 0

			_, err := railway.Sequentially(
				func() (ccv3.Warnings, error) { i = 1; return ccv3.Warnings{}, nil },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{}, errors.New("error 1") },
			)

			Expect(err).To(Equal(errors.New("error 1")))
			Expect(i).To(Equal(1))
		})
	})

	Describe("with warnings", func() {
		It("returns empty warnings when nothing runs", func() {
			Expect(railway.Sequentially()).To(BeEmpty())
		})

		It("returns combines warnings on successes", func() {
			warnings, _ := railway.Sequentially(
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 1"}, nil },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 2"}, nil },
			)

			Expect(warnings).To(ConsistOf("warning 1", "warning 2"))
		})

		It("returns warnings that happened before an error", func() {
			warnings, _ := railway.Sequentially(
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 1"}, nil },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{}, errors.New("error 1") },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 3"}, nil },
			)

			Expect(warnings).To(ConsistOf("warning 1"))
		})

		It("includes warnings from the failed track", func() {
			warnings, _ := railway.Sequentially(
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 1"}, nil },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 2"}, errors.New("error 1") },
				func() (ccv3.Warnings, error) { return ccv3.Warnings{"warning 3"}, nil },
			)

			Expect(warnings).To(ConsistOf("warning 1", "warning 2"))
		})
	})
})
