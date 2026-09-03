package shared_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	. "code.cloudfoundry.org/cli/v8/command/v7/shared"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("WaitForResult", func() {
	var (
		testUI    *ui.UI
		stream    chan v7action.PollJobEvent
		completed bool
		err       error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
	})

	When("the stream is nil (synchronous operation)", func() {
		It("reports completion with no error", func() {
			completed, err = WaitForResult(nil, testUI, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeTrue())
		})
	})

	When("not waiting for completion and the job starts polling", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"a warning"}}
			// channel intentionally left open
		})

		It("returns not-completed and displays the warning", func() {
			completed, err = WaitForResult(stream, testUI, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeFalse())
			Expect(testUI.Err).To(Say("a warning"))
		})
	})

	When("an event carries an error", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobFailed, Err: errors.New("boom")}
			close(s)
		})

		It("returns the error and not-completed", func() {
			completed, err = WaitForResult(stream, testUI, false)
			Expect(err).To(MatchError("boom"))
			Expect(completed).To(BeFalse())
		})
	})

	When("the job completes", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobComplete}
			close(s)
		})

		It("returns completed with no error", func() {
			completed, err = WaitForResult(stream, testUI, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeTrue())
		})
	})

	Describe("duplicate warning suppression", func() {
		When("the same warning is re-sent on later polls", func() {
			BeforeEach(func() {
				s := make(chan v7action.PollJobEvent, 3)
				stream = s
				s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"still in progress"}}
				s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"still in progress"}}
				s <- v7action.PollJobEvent{State: v7action.JobComplete, Warnings: v7action.Warnings{"still in progress"}}
				close(s)
			})

			It("prints the warning only once", func() {
				completed, err = WaitForResult(stream, testUI, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(completed).To(BeTrue())
				Expect(testUI.Err).To(Say("still in progress"))
				Expect(testUI.Err).NotTo(Say("still in progress"))
			})
		})

		When("distinct warnings arrive", func() {
			BeforeEach(func() {
				s := make(chan v7action.PollJobEvent, 2)
				stream = s
				s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"first"}}
				s <- v7action.PollJobEvent{State: v7action.JobComplete, Warnings: v7action.Warnings{"second"}}
				close(s)
			})

			It("prints each distinct warning, preserving order", func() {
				completed, err = WaitForResult(stream, testUI, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("first"))
				Expect(testUI.Err).To(Say("second"))
			})
		})

		When("a warning recurs after a distinct warning intervened", func() {
			BeforeEach(func() {
				s := make(chan v7action.PollJobEvent, 3)
				stream = s
				s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"persisted"}}
				s <- v7action.PollJobEvent{State: v7action.JobPolling, Warnings: v7action.Warnings{"persisted", "changed"}}
				s <- v7action.PollJobEvent{State: v7action.JobComplete, Warnings: v7action.Warnings{"persisted"}}
				close(s)
			})

			It("prints each distinct warning only once for the whole operation", func() {
				completed, err = WaitForResult(stream, testUI, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("persisted"))
				Expect(testUI.Err).To(Say("changed"))
				Expect(testUI.Err).NotTo(Say("persisted"))
			})
		})
	})
})
