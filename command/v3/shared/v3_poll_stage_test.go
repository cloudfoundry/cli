package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v3/shared"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock/fakeclock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("V3PollStage", func() {
	var (
		appName               string
		returnedDropletGUID   string
		executeErr            error
		testUI                *ui.UI
		buildStream           chan v3action.Build
		warningsStream        chan v3action.Warnings
		errStream             chan error
		logStream             chan *v3action.LogMessage
		logErrStream          chan error
		fakeClock             *fakeclock.FakeClock
		closeStreams          func()
		executePollStage      func(func())
		finishedWritingEvents chan bool
		finishedClosing       chan bool
	)

	closeStreams = func() {
		close(errStream)
		close(warningsStream)
		close(buildStream)
		finishedClosing <- true
	}

	executePollStage = func(codeAssertions func()) {
		returnedDropletGUID, executeErr = PollStage(
			appName,
			buildStream,
			warningsStream,
			errStream,
			logStream,
			logErrStream,
			testUI,
			fakeClock,
			15*time.Minute)
		codeAssertions()
		Eventually(finishedClosing).Should(Receive(Equal(true)))
	}

	BeforeEach(func() {
		// reset assertion variables
		executeErr = nil
		returnedDropletGUID = ""

		// create new channels
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeClock = fakeclock.NewFakeClock(time.Now())
		buildStream = make(chan v3action.Build)
		warningsStream = make(chan v3action.Warnings)
		errStream = make(chan error)
		logStream = make(chan *v3action.LogMessage)
		logErrStream = make(chan error)
		appName = "some-app"

		finishedWritingEvents = make(chan bool)
		finishedClosing = make(chan bool)

		// wait for all events to be written before closing channels
		go func() {
			defer GinkgoRecover()

			Eventually(finishedWritingEvents).Should(Receive(Equal(true)))
			closeStreams()
		}()
	})

	Context("when the build stream contains a droplet GUID", func() {
		BeforeEach(func() {
			go func() {
				defer GinkgoRecover()

				buildStream <- v3action.Build{Droplet: ccv3.Droplet{GUID: "droplet-guid"}}
				finishedWritingEvents <- true
			}()
		})

		It("returns the droplet GUID", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDropletGUID).To(Equal("droplet-guid"))
			})

			Eventually(testUI.Out).Should(Say("droplet: droplet-guid"))
		})
	})

	Context("when the warnings stream contains warnings", func() {
		BeforeEach(func() {
			go func() {
				defer GinkgoRecover()

				warningsStream <- v3action.Warnings{"warning-1", "warning-2"}
				finishedWritingEvents <- true
			}()
		})

		It("displays the warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDropletGUID).To(BeEmpty())
			})

			Eventually(testUI.Err).Should(Say("warning-1"))
			Eventually(testUI.Err).Should(Say("warning-2"))
		})
	})

	Context("when the log stream contains a log message", func() {
		Context("and the message is a staging message", func() {
			BeforeEach(func() {
				go func() {
					defer GinkgoRecover()

					logStream <- v3action.NewLogMessage("some-log-message", 1, time.Now(), v3action.StagingLog, "1")
					finishedWritingEvents <- true
				}()
			})

			It("prints the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDropletGUID).To(BeEmpty())
				})
				Eventually(testUI.Out).Should(Say("some-log-message"))
			})
		})

		Context("and the message is not a staging message", func() {
			BeforeEach(func() {
				go func() {
					defer GinkgoRecover()

					logStream <- v3action.NewLogMessage("some-log-message", 1, time.Now(), "RUN", "1")
					finishedWritingEvents <- true
				}()
			})

			It("ignores the log message", func() {
				executePollStage(func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedDropletGUID).To(BeEmpty())
				})
				Consistently(testUI.Out).ShouldNot(Say("some-log-message"))
			})
		})
	})

	Context("when the error stream contains an error", func() {
		BeforeEach(func() {
			go func() {
				defer GinkgoRecover()

				errStream <- errors.New("some error")
				finishedWritingEvents <- true
			}()
		})

		It("returns the error without waiting for streams to be closed", func() {
			executePollStage(func() {
				Expect(executeErr).To(MatchError("some error"))
				Expect(returnedDropletGUID).To(BeEmpty())
			})
			Consistently(testUI.Out).ShouldNot(Say("droplet: droplet-guid"))
		})
	})

	Context("when the log error stream contains errors", func() {
		BeforeEach(func() {
			go func() {
				defer GinkgoRecover()

				logErrStream <- errors.New("some-log-error")
				finishedWritingEvents <- true
			}()
		})

		It("displays the log errors as warnings", func() {
			executePollStage(func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedDropletGUID).To(BeEmpty())
			})
			Eventually(testUI.Err).Should(Say("some-log-error"))
		})
	})

	XContext("when the staging timeout has expired", func() {
		It("exits with an error", func() {
			// Consistently(done).ShouldNot(Receive())
			fakeClock.WaitForWatcherAndIncrement(15*time.Minute - time.Nanosecond)
			// Consistently(done).ShouldNot(Receive())

			fakeClock.Increment(time.Duration(time.Nanosecond))
			// Eventually(done).Should(Receive())

			executePollStage(func() {
				Expect(executeErr).To(MatchError(translatableerror.StagingTimeoutError{
					AppName: "some-app",
					Timeout: 15 * time.Minute}))
				Expect(returnedDropletGUID).To(BeEmpty())
			})
		})
	})
})
