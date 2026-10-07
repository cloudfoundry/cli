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
		jobGUID   string
		err       error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
	})

	When("the stream is nil (synchronous operation)", func() {
		It("reports completion with no error and no job GUID", func() {
			completed, jobGUID, err = WaitForResult(nil, testUI, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeTrue())
			Expect(jobGUID).To(BeEmpty())
		})
	})

	When("not waiting for completion and the job starts polling", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobPolling, JobGUID: "the-job-guid", Warnings: v7action.Warnings{"a warning"}}
			// channel intentionally left open
		})

		It("returns not-completed and hands back the observed job GUID", func() {
			completed, jobGUID, err = WaitForResult(stream, testUI, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeFalse())
			Expect(jobGUID).To(Equal("the-job-guid"))
			Expect(testUI.Err).To(Say("a warning"))
		})
	})

	When("an event carries an error", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobFailed, Err: errors.New("boom"), JobGUID: "err-job-guid"}
			close(s)
		})

		It("returns the error, not-completed, and the observed GUID", func() {
			completed, jobGUID, err = WaitForResult(stream, testUI, false)
			Expect(err).To(MatchError("boom"))
			Expect(completed).To(BeFalse())
			Expect(jobGUID).To(Equal("err-job-guid"))
		})
	})

	When("the job completes", func() {
		BeforeEach(func() {
			s := make(chan v7action.PollJobEvent, 1)
			stream = s
			s <- v7action.PollJobEvent{State: v7action.JobComplete, JobGUID: "complete-guid"}
			close(s)
		})

		It("returns completed with no error and the GUID", func() {
			completed, jobGUID, err = WaitForResult(stream, testUI, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(completed).To(BeTrue())
			Expect(jobGUID).To(Equal("complete-guid"))
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
				completed, jobGUID, err = WaitForResult(stream, testUI, true)
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
				completed, jobGUID, err = WaitForResult(stream, testUI, true)
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
				completed, jobGUID, err = WaitForResult(stream, testUI, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("persisted"))
				Expect(testUI.Err).To(Say("changed"))
				Expect(testUI.Err).NotTo(Say("persisted"))
			})
		})
	})
})

var _ = Describe("DisplayJobHint", func() {
	var testUI *ui.UI

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
	})

	When("a job GUID is given", func() {
		It("prints a hint naming the job", func() {
			DisplayJobHint(testUI, "the-job-guid")
			Expect(testUI.Out).To(Say(`Job \(the-job-guid\) is being processed\.`))
		})
	})

	When("the job GUID is empty", func() {
		It("prints nothing", func() {
			DisplayJobHint(testUI, "")
			Expect(testUI.Out).NotTo(Say("is being processed"))
		})
	})
})
