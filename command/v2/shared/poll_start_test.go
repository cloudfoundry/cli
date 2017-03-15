package shared_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2/shared"

	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Poll Start", func() {
	var (
		testUI      *ui.UI
		fakeConfig  *commandfakes.FakeConfig
		messages    chan *v2action.LogMessage
		logErrs     chan error
		appStarting chan bool
		apiWarnings chan string
		apiErrs     chan error
		err         error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeConfig.BinaryNameReturns("FiveThirtyEight")

		messages = make(chan *v2action.LogMessage)
		logErrs = make(chan error)
		appStarting = make(chan bool)
		apiWarnings = make(chan string)
		apiErrs = make(chan error)

		err = errors.New("This should never occur.")

		go func() {
			err = PollStart(testUI, fakeConfig, messages, logErrs, appStarting, apiWarnings, apiErrs)
		}()
	})

	Context("when no API errors appear", func() {
		It("passes and exits with no errors", func() {
			appStarting <- true
			logErrs <- v2action.NOAATimeoutError{}
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
			close(appStarting)
			apiWarnings <- "some other warning"
			close(apiWarnings)
			close(apiErrs)

			Eventually(testUI.Out).Should(Say("\nWaiting for app to start..."))
			Consistently(testUI.Out).ShouldNot(Say("\nWaiting for app to start..."))
			Eventually(testUI.Err).Should(Say("timeout connecting to log server, no log will be shown"))
			Eventually(testUI.Err).Should(Say("some warning"))
			Eventually(testUI.Err).Should(Say("some logErrhea"))
			Eventually(testUI.Out).Should(Say("some log message"))
			Consistently(testUI.Out).ShouldNot(Say("some other log messsage"))
			Eventually(testUI.Err).Should(Say("some other warning"))
			Eventually(func() error { return err }).Should(BeNil())
		})

		It("passes and exits with no errors or duplicated output", func() {
			appStarting <- true
			appStarting <- false
			close(appStarting)
			close(apiWarnings)
			close(apiErrs)

			Eventually(testUI.Out).Should(Say("\nWaiting for app to start..."))
			Consistently(testUI.Out).ShouldNot(Say("\nWaiting for app to start..."))
			Eventually(func() error { return err }).Should(BeNil())
		})
	})

	Context("when there are API errors", func() {
		It("StagingFailedError", func() {
			apiErrs <- v2action.StagingFailedError{
				Reason: "some staging failure reason",
			}
			Eventually(func() error { return err }).Should(MatchError(StagingFailedError{
				Message:    "some staging failure reason",
				BinaryName: "FiveThirtyEight"}))
		})

		It("StagingTimeoutError", func() {
			apiErrs <- v2action.StagingTimeoutError{
				Name:    "some staging timeout name",
				Timeout: time.Second,
			}
			Eventually(func() error { return err }).Should(MatchError(StagingTimeoutError{
				AppName: "some staging timeout name",
				Timeout: time.Second}))
		})

		It("ApplicationInstanceCrashedError", func() {
			apiErrs <- v2action.ApplicationInstanceCrashedError{
				Name: "some application crashed name",
			}
			Eventually(func() error { return err }).Should(MatchError(UnsuccessfulStartError{
				AppName:    "some application crashed name",
				BinaryName: "FiveThirtyEight"}))
		})

		It("ApplicationInstanceFlappingError", func() {
			apiErrs <- v2action.ApplicationInstanceFlappingError{
				Name: "some application flapping name",
			}
			Eventually(func() error { return err }).Should(MatchError(UnsuccessfulStartError{
				AppName:    "some application flapping name",
				BinaryName: "FiveThirtyEight"}))
		})

		It("StartupTimeoutError", func() {
			apiErrs <- v2action.StartupTimeoutError{
				Name: "some application timeout name",
			}
			Eventually(func() error { return err }).Should(MatchError(StartupTimeoutError{
				AppName:    "some application timeout name",
				BinaryName: "FiveThirtyEight"}))
		})

		It("any other error", func() {
			apiErrs <- v2action.HTTPHealthCheckInvalidError{}
			Eventually(func() error { return err }).Should(MatchError(HTTPHealthCheckInvalidError{}))
		})
	})
})
