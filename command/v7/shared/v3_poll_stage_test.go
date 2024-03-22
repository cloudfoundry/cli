package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("V3PollStage", func() {
	var (
		returnedDroplet       resources.Droplet
		executeErr            error
		testUI                *ui.UI
		dropletStream         chan resources.Droplet
		warningsStream        chan v7action.Warnings
		errStream             chan error
		logStream             chan sharedaction.LogMessage
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
		returnedDroplet = resources.Droplet{}

		// create new channels
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		dropletStream = make(chan resources.Droplet)
		warningsStream = make(chan v7action.Warnings)
		errStream = make(chan error)
		logStream = make(chan sharedaction.LogMessage)
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

	When("the droplet stream contains a droplet GUID", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				dropletStream <- resources.Droplet{GUID: "droplet-guid"}
			})
		})

		It("returns the droplet GUID", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet.GUID).To(Equal("droplet-guid"))
			})
		})
	})

	When("the warnings stream contains warnings", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				warningsStream <- v7action.Warnings{"warning-1", "warning-2"}
			})
		})

		It("displays the warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet).To(Equal(resources.Droplet{}))
			})

			Eventually(testUI.Err).Should(Say("warning-1"))
			Eventually(testUI.Err).Should(Say("warning-2"))
		})
	})

	When("the log stream contains a log message", func() {
		Context("and the message is a staging message", func() {
			BeforeEach(func() {
				writeEventsAsync(func() {
					logStream <- *sharedaction.NewLogMessage("some-log-message", "OUT", time.Now(), sharedaction.StagingLog, "1")
				})
			})

			It("prints the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDroplet).To(Equal(resources.Droplet{}))
				})
				Eventually(testUI.Out).Should(Say("some-log-message"))
			})
		})

		Context("and the message is not a staging message", func() {
			BeforeEach(func() {
				writeEventsAsync(func() {
					logStream <- *sharedaction.NewLogMessage("some-log-message", "OUT", time.Now(), "RUN", "1")
				})
			})

			It("ignores the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDroplet).To(Equal(resources.Droplet{}))
				})
				Consistently(testUI.Out).ShouldNot(Say("some-log-message"))
			})
		})
	})

	When("the error stream contains an error", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				errStream <- errors.New("some error")
			})
		})

		It("returns the error without waiting for streams to be closed", func() {
			executePollStage(func() {
				Expect(executeErr).To(MatchError("some error"))
				Expect(returnedDroplet).To(Equal(resources.Droplet{}))
			})
		})
	})

	When("the log error stream contains errors", func() {
		BeforeEach(func() {
			writeEventsAsync(func() {
				logErrStream <- actionerror.LogCacheTimeoutError{}
				logErrStream <- errors.New("some-log-error")
			})
		})

		It("displays the log errors as warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDroplet).To(Equal(resources.Droplet{}))
			})
			Eventually(testUI.Err).Should(Say("timeout connecting to log server, no log will be shown"))
			Eventually(testUI.Err).Should(Say("some-log-error"))
		})
	})
})
