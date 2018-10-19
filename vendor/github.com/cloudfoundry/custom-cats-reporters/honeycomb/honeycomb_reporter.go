package honeycomb

import (
	"strconv"
	"strings"

	"github.com/cloudfoundry/custom-cats-reporters/honeycomb/client"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type SpecEvent struct {
	Description           string
	State                 string
	FailureMessage        string
	FailureLocation       string
	FailureOutput         string
	ComponentCodeLocation string
	ComponentType         string
	RunTimeInSeconds      string
}

type honeyCombReporter struct {
	client     client.Client
	globalTags map[string]interface{}
	customTags map[string]interface{}
}

func New(client client.Client) honeyCombReporter {
	return honeyCombReporter{client: client}
}

func (hr honeyCombReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	specEvent := SpecEvent{
		State:       getTestState(specSummary.State),
		Description: createTestDescription(specSummary.ComponentTexts),
	}

	if specSummary.State == types.SpecStateFailed {
		specEvent.FailureMessage = specSummary.Failure.Message
		specEvent.ComponentCodeLocation = specSummary.Failure.ComponentCodeLocation.String()
		specEvent.FailureLocation = specSummary.Failure.Location.String()
		specEvent.FailureOutput = specSummary.CapturedOutput
		specEvent.ComponentType = getComponentType(specSummary.Failure.ComponentType)
	}
	if specSummary.State == types.SpecStateFailed || specSummary.State == types.SpecStatePassed {
		specEvent.RunTimeInSeconds = strconv.Itoa(int(specSummary.RunTime.Seconds()))
	}
	// intentionally drop all errors to satisfy reporter interface
	// and avoid unnecessary noise when an event cannot be sent to honeycomb
	hr.client.SendEvent(specEvent, hr.globalTags, hr.customTags)
}

func (hr honeyCombReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	// intentionally drop all errors to satisfy reporter interface
	// and avoid unnecessary noise when an event cannot be sent to honeycomb
	hr.client.SendEvent(*summary, hr.globalTags, hr.customTags)
}

func (hr honeyCombReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	specEvent := SpecEvent{
		State:         getTestState(setupSummary.State),
		ComponentType: getComponentType(setupSummary.ComponentType),
	}

	if setupSummary.State == types.SpecStateFailed {
		specEvent.FailureMessage = setupSummary.Failure.Message
		specEvent.ComponentCodeLocation = setupSummary.Failure.ComponentCodeLocation.String()
		specEvent.FailureLocation = setupSummary.Failure.Location.String()
		specEvent.FailureOutput = setupSummary.CapturedOutput
	}

	hr.client.SendEvent(specEvent, hr.globalTags, hr.customTags)
}

func (hr honeyCombReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (hr honeyCombReporter) SpecWillRun(specSummary *types.SpecSummary)        {}
func (hr honeyCombReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (hr *honeyCombReporter) SetGlobalTags(globalTags map[string]interface{}) {
	hr.globalTags = globalTags
}

func (hr *honeyCombReporter) SetCustomTags(customTags map[string]interface{}) {
	hr.customTags = customTags
}

func getTestState(state types.SpecState) string {
	switch state {
	case types.SpecStatePassed:
		return "passed"
	case types.SpecStateFailed:
		return "failed"
	case types.SpecStatePending:
		return "pending"
	case types.SpecStateSkipped:
		return "skipped"
	case types.SpecStatePanicked:
		return "panicked"
	case types.SpecStateTimedOut:
		return "timedOut"
	case types.SpecStateInvalid:
		return "invalid"
	default:
		panic("unknown spec state")
	}
}

func getComponentType(thingie types.SpecComponentType) string {
	switch thingie {
	case types.SpecComponentTypeInvalid:
		return "invalid"
	case types.SpecComponentTypeContainer:
		return "container"
	case types.SpecComponentTypeBeforeSuite:
		return "beforeSuite"
	case types.SpecComponentTypeAfterSuite:
		return "afterSuite"
	case types.SpecComponentTypeBeforeEach:
		return "beforeEach"
	case types.SpecComponentTypeJustBeforeEach:
		return "justBeforeEach"
	case types.SpecComponentTypeAfterEach:
		return "afterEach"
	case types.SpecComponentTypeIt:
		return "it"
	case types.SpecComponentTypeMeasure:
		return "measure"
	default:
		panic("unknown spec component")
	}
}

func createTestDescription(componentTexts []string) string {
	return strings.Join(componentTexts, " | ")
}
