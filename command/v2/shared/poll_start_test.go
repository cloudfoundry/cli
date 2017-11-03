package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2/shared"

	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Poll Start", func() {
	var (
		testUI      *ui.UI
		fakeConfig  *commandfakes.FakeConfig
		messages    chan *v2action.LogMessage
		logErrs     chan error
		appState    chan v2action.ApplicationStateChange
		apiWarnings chan string
		apiErrs     chan error
		err         error
		block       chan bool
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("FiveThirtyEight")

		messages = make(chan *v2action.LogMessage)
		logErrs = make(chan error)
		appState = make(chan v2action.ApplicationStateChange)
		apiWarnings = make(chan string)
		apiErrs = make(chan error)
		block = make(chan bool)

		err = errors.New("This should never occur.")
	})

	JustBeforeEach(func() {
		go func() {
			err = PollStart(testUI, fakeConfig, messages, logErrs, appState, apiWarnings, apiErrs)
			close(block)
		}()
	})

	Context("when no API errors appear", func() {
		It("passes and exits with no errors", func() {
			appState <- v2action.ApplicationStateStopping
			appState <- v2action.ApplicationStateStaging
			appState <- v2action.ApplicationStateStarting
			logErrs <- actionerror.NOAATimeoutError{}
			apiWarnings <- "some warning"
			logErrs <- errors.New("some logErrhea")
			messages <- v2action.NewLogMessage(
				"some log message",
				1,
				time.Unix(0, 0),
				"STG",
				"some source instance")
			messages <- v2action.NewLogMessage(
				"some other log message",
				1,
				time.Unix(0, 0),
				"APP",
				"some other source instance")
			close(appState)
			apiWarnings <- "some other warning"
			close(apiWarnings)
			close(apiErrs)

			Eventually(testUI.Out).Should(Say("\nStopping app..."))
			Eventually(testUI.Out).Should(Say("\nStaging app and tracing logs..."))
			Eventually(testUI.Out).Should(Say("\nWaiting for app to start..."))
			Eventually(testUI.Err).Should(Say("timeout connecting to log server, no log will be shown"))
			Eventually(testUI.Err).Should(Say("some warning"))
			Eventually(testUI.Err).Should(Say("some logErrhea"))
			Eventually(testUI.Out).Should(Say("some log message"))
			Consistently(testUI.Out).ShouldNot(Say("some other log messsage"))
			Eventually(testUI.Err).Should(Say("some other warning"))
			Eventually(block).Should(BeClosed())
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when state channel is not set", func() {
			BeforeEach(func() {
				appState = nil
			})

			It("does not wait for it", func() {
				close(apiWarnings)
				close(apiErrs)

				Eventually(block).Should(BeClosed())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	DescribeTable("API Errors",
		func(apiErr error, expectedErr error) {
			apiErrs <- apiErr
			Eventually(block).Should(BeClosed())
			Expect(err).To(MatchError(expectedErr))
		},

		Entry("StagingFailedNoAppDetectedError",
			actionerror.StagingFailedNoAppDetectedError{
				Reason: "some staging failure reason",
			},
			translatableerror.StagingFailedNoAppDetectedError{
				Message:    "some staging failure reason",
				BinaryName: "FiveThirtyEight",
			},
		),

		Entry("StagingFailedError",
			actionerror.StagingFailedError{
				Reason: "some staging failure reason",
			},
			translatableerror.StagingFailedError{
				Message: "some staging failure reason",
			},
		),

		Entry("StagingTimeoutError",
			actionerror.StagingTimeoutError{
				AppName: "some staging timeout name",
				Timeout: time.Second,
			},
			translatableerror.StagingTimeoutError{
				AppName: "some staging timeout name",
				Timeout: time.Second,
			},
		),

		Entry("ApplicationInstanceCrashedError",
			actionerror.ApplicationInstanceCrashedError{
				Name: "some application crashed name",
			},
			translatableerror.UnsuccessfulStartError{
				AppName:    "some application crashed name",
				BinaryName: "FiveThirtyEight",
			},
		),

		Entry("ApplicationInstanceFlappingError",
			actionerror.ApplicationInstanceFlappingError{
				Name: "some application flapping name",
			},
			translatableerror.UnsuccessfulStartError{
				AppName:    "some application flapping name",
				BinaryName: "FiveThirtyEight",
			},
		),

		Entry("StartupTimeoutError",
			actionerror.StartupTimeoutError{
				Name: "some application timeout name",
			},
			translatableerror.StartupTimeoutError{
				AppName:    "some application timeout name",
				BinaryName: "FiveThirtyEight",
			},
		),

		Entry("any other error",
			actionerror.HTTPHealthCheckInvalidError{},
			actionerror.HTTPHealthCheckInvalidError{},
		),
	)
})
