package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

// RealtimeReporter is a Ginkgo reporter that streams test events in real-time
type RealtimeReporter struct {
	writer      io.Writer
	suite       TestSuite
	specs       []TestSpec
	currentSpec *TestSpec
	mu          sync.Mutex
	startTime   time.Time
}

// TestSuite represents the overall test suite
type TestSuite struct {
	Name              string    `json:"name"`
	TotalSpecs        int       `json:"total_specs"`
	CompletedSpecs    int       `json:"completed_specs"`
	PassedSpecs       int       `json:"passed_specs"`
	FailedSpecs       int       `json:"failed_specs"`
	PendingSpecs      int       `json:"pending_specs"`
	SkippedSpecs      int       `json:"skipped_specs"`
	RunningTime       float64   `json:"running_time"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"start_time"`
	EstimatedTimeLeft float64   `json:"estimated_time_left"`
}

// TestSpec represents a single test specification
type TestSpec struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"` // running, passed, failed, pending, skipped
	Duration   float64   `json:"duration"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	FileName   string    `json:"file_name"`
	LineNumber int       `json:"line_number"`
	Error      string    `json:"error,omitempty"`
	FullText   string    `json:"full_text"`
}

// TestEvent represents a real-time event
type TestEvent struct {
	Type      string      `json:"type"` // suite_start, spec_start, spec_end, suite_end
	Timestamp time.Time   `json:"timestamp"`
	Suite     *TestSuite  `json:"suite,omitempty"`
	Spec      *TestSpec   `json:"spec,omitempty"`
	Message   string      `json:"message,omitempty"`
}

// NewRealtimeReporter creates a new real-time reporter
func NewRealtimeReporter(outputPath string) (*RealtimeReporter, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	return &RealtimeReporter{
		writer: file,
		specs:  make([]TestSpec, 0),
	}, nil
}

// SpecSuiteWillBegin is called at the start of the test suite
func (r *RealtimeReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.startTime = time.Now()
	r.suite = TestSuite{
		Name:       summary.SuiteDescription,
		TotalSpecs: summary.NumberOfSpecsThatWillBeRun,
		Status:     "running",
		StartTime:  r.startTime,
	}

	event := TestEvent{
		Type:      "suite_start",
		Timestamp: time.Now(),
		Suite:     &r.suite,
		Message:   fmt.Sprintf("Starting suite: %s (%d specs)", r.suite.Name, r.suite.TotalSpecs),
	}

	r.emitEvent(event)
}

// BeforeSuiteDidRun is called before suite setup
func (r *RealtimeReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	// Optional: track setup time
}

// SpecWillRun is called before each spec
func (r *RealtimeReporter) SpecWillRun(specSummary *types.SpecSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()

	spec := TestSpec{
		Name:       specSummary.ComponentTexts[len(specSummary.ComponentTexts)-1],
		FullText:   fmt.Sprintf("%v", specSummary.ComponentTexts),
		Status:     "running",
		StartTime:  time.Now(),
		FileName:   specSummary.ComponentCodeLocations[0].FileName,
		LineNumber: specSummary.ComponentCodeLocations[0].LineNumber,
	}

	r.currentSpec = &spec

	event := TestEvent{
		Type:      "spec_start",
		Timestamp: time.Now(),
		Spec:      &spec,
		Message:   fmt.Sprintf("Running: %s", spec.Name),
	}

	r.emitEvent(event)
}

// SpecDidComplete is called after each spec
func (r *RealtimeReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentSpec == nil {
		return
	}

	r.currentSpec.EndTime = time.Now()
	r.currentSpec.Duration = r.currentSpec.EndTime.Sub(r.currentSpec.StartTime).Seconds()

	switch {
	case specSummary.State == types.SpecStatePassed:
		r.currentSpec.Status = "passed"
		r.suite.PassedSpecs++
	case specSummary.State == types.SpecStateFailed:
		r.currentSpec.Status = "failed"
		r.suite.FailedSpecs++
		if specSummary.Failure.Message != "" {
			r.currentSpec.Error = specSummary.Failure.Message
		}
	case specSummary.State == types.SpecStatePending:
		r.currentSpec.Status = "pending"
		r.suite.PendingSpecs++
	case specSummary.State == types.SpecStateSkipped:
		r.currentSpec.Status = "skipped"
		r.suite.SkippedSpecs++
	}

	r.suite.CompletedSpecs++
	r.suite.RunningTime = time.Since(r.startTime).Seconds()

	// Calculate estimated time left
	if r.suite.CompletedSpecs > 0 {
		avgTimePerSpec := r.suite.RunningTime / float64(r.suite.CompletedSpecs)
		remaining := r.suite.TotalSpecs - r.suite.CompletedSpecs
		r.suite.EstimatedTimeLeft = avgTimePerSpec * float64(remaining)
	}

	r.specs = append(r.specs, *r.currentSpec)

	event := TestEvent{
		Type:      "spec_end",
		Timestamp: time.Now(),
		Suite:     &r.suite,
		Spec:      r.currentSpec,
	}

	r.emitEvent(event)
	r.currentSpec = nil
}

// AfterSuiteDidRun is called after suite teardown
func (r *RealtimeReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	// Optional: track teardown time
}

// SpecSuiteDidEnd is called at the end of the test suite
func (r *RealtimeReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.suite.RunningTime = time.Since(r.startTime).Seconds()

	if summary.SuiteSucceeded {
		r.suite.Status = "passed"
	} else {
		r.suite.Status = "failed"
	}

	event := TestEvent{
		Type:      "suite_end",
		Timestamp: time.Now(),
		Suite:     &r.suite,
		Message:   fmt.Sprintf("Suite completed: %d passed, %d failed", r.suite.PassedSpecs, r.suite.FailedSpecs),
	}

	r.emitEvent(event)
}

// emitEvent writes an event to the output stream
func (r *RealtimeReporter) emitEvent(event TestEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintf(r.writer, "%s\n", data)
}

// GetSummary returns the current test summary
func (r *RealtimeReporter) GetSummary() TestSuite {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.suite
}

// GetSpecs returns all completed specs
func (r *RealtimeReporter) GetSpecs() []TestSpec {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]TestSpec(nil), r.specs...)
}
