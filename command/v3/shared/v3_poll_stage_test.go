package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v3action"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("V3PollStage", func() {
	var (
		returnedDroplet       v3action.Droplet
		executeErr            error
		testUI                *ui.UI
		dropletStream         chan v3action.Droplet
		warningsStream        chan v3action.Warnings
		errStream             chan error
		logStream             chan *v3action.LogMessage
		logErrStream          chan error
		closeStreams          func()
		writeEventsAsync      func(func())
		executePollStage      func(func())
		finishedWritingEvents chan bool
		finishedClosing       chan bool
	)

	closeStreams = func() {
		close(errStream)
		close(warningsStream)
		close(dropletStream)
		finishedClosing <- true
	}

	writeEventsAsync = func(writeEvents func()) {
		go func() {
			defer GinkgoRecover()
			writeEvents()
			finishedWritingEvents <- true
		}()
	}

	executePollStage = func(codeAssertions func()) {
		returnedDroplet, executeErr = PollStage(
			dropletStream,
			warningsStream,
			errStream,
			logStream,
			logErrStream,
			testUI)
		codeAssertions()
		Eventually(finishedClosing).Should(Receive(Equal(true)))
	}

	BeforeEach(func() {
		// reset assertion variables
		executeErr = nil
		returnedDroplet = v3action.Droplet{}

		// create new channels
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		dropletStream = make(chan v3action.Droplet)
		warningsStream = make(chan v3action.Warnings)
		errStream = make(chan error)
		logStream = make(chan *v3action.LogMessage)
		logErrStream = make(chan error)

		finishedWritingEvents = make(chan bool)
		finishedClosing = make(chan bool)

		// wait for all events to be written before closing channels
		go func() {
			defer GinkgoRecover()

			Eventually(finishedWritingEvents).Should(Receive(Equal(true)))
			closeStreams()
		}()
	})

	Context("when the droplet stream contains a droplet GUID", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				dropletStream <- v3action.Droplet{GUID: "droplet-guid"}
			})
		})

		It("returns the droplet GUID", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet.GUID).To(Equal("droplet-guid"))
			})
		})
	})

	Context("when the warnings stream contains warnings", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				warningsStream <- v3action.Warnings{"warning-1", "warning-2"}
			})
		})

		It("displays the warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet).To(Equal(v3action.Droplet{}))
			})

			Eventually(testUI.Err).Should(Say("warning-1"))
			Eventually(testUI.Err).Should(Say("warning-2"))
		})
	})

	Context("when the log stream contains a log message", func() {
		Context("and the message is a staging message", func() {
			BeforeEach(func() {
				writeEventsAsync(func() {
					logStream <- v3action.NewLogMessage("some-log-message", 1, time.Now(), v3action.StagingLog, "1")
				})
			})

			It("prints the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDroplet).To(Equal(v3action.Droplet{}))
				})
				Eventually(testUI.Out).Should(Say("some-log-message"))
			})
		})

		Context("and the message is not a staging message", func() {
			BeforeEach(func() {
				writeEventsAsync(func() {
					logStream <- v3action.NewLogMessage("some-log-message", 1, time.Now(), "RUN", "1")
				})
			})

			It("ignores the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDroplet).To(Equal(v3action.Droplet{}))
				})
				Consistently(testUI.Out).ShouldNot(Say("some-log-message"))
			})
		})
	})

	Context("when the error stream contains an error", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				errStream <- errors.New("some error")
			})
		})

		It("returns the error without waiting for streams to be closed", func() {
			executePollStage(func() {
				Expect(executeErr).To(MatchError("some error"))
				Expect(returnedDroplet).To(Equal(v3action.Droplet{}))
			})
		})
	})

	Context("when the log error stream contains errors", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				logErrStream <- errors.New("some-log-error")
			})
		})

		It("displays the log errors as warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet).To(Equal(v3action.Droplet{}))
			})
			Eventually(testUI.Err).Should(Say("some-log-error"))
		})
	})
})
