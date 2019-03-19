package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

const (
	PRBuilderOutputEnvVar = "PR_BUILDER_OUTPUT_DIR"
)

func GetPRBuilderReporter() ginkgo.Reporter {
	outputDir := os.Getenv(PRBuilderOutputEnvVar)

	if outputDir == "" {
		return nil
	}

	prBuilderReporter := NewPRBuilderReporter(outputDir)
	return prBuilderReporter
}

type PRBuilderReporter struct {
	outputFile *os.File
}

func NewPRBuilderReporter(outputDir string) *PRBuilderReporter {
	outputFile := filepath.Join(outputDir, strconv.Itoa(ginkgo.GinkgoParallelNode()))

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	reporter := &PRBuilderReporter{outputFile: f}
	return reporter
}

func (reporter *PRBuilderReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if specSummary.Failed() {
		msg := trimmedLocation(specSummary.Failure.Location)
		_, err := reporter.outputFile.WriteString(msg + "\n")
		if err != nil {
			panic(err)
		}
	}
}

// WriteFailureSummary aggregates test failures from all parallel nodes, sorts
// them, and writes the result to a file.
func WriteFailureSummary(outputRoot, filename string) {
	outfile, err := os.Create(filepath.Join(outputRoot, filename))
	failureSummaries := make([]string, 0)
	if err != nil {
		panic(err)
	}
	err = filepath.Walk(outputRoot, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.Contains(path, "summary") {
			return nil
		}
		defer os.Remove(path)
		allFailures, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		failures := strings.Split(string(allFailures), "\n")
		failureSummaries = append(failureSummaries, failures...)
		return nil
	})
	if err != nil {
		panic(err)
	}
	sort.Strings(failureSummaries)
	anyFailures := false
	var previousLine string
	for _, line := range failureSummaries {
		if line != "" && line != previousLine {
			anyFailures = true
			previousLine = line
			fmt.Fprintln(outfile, line)
		}
	}
	if !anyFailures {
		err = os.Remove(filepath.Join(outputRoot, filename))
	}
	if err != nil {
		panic(err)
	}
}

func trimmedLocation(location types.CodeLocation) string {
	splits := strings.Split(location.String(), "/cli/")
	return strings.Join(splits[1:], "")
}

// unused members of ginkgo reporter interface

func (reporter *PRBuilderReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

func (reporter *PRBuilderReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {}

func (reporter *PRBuilderReporter) SpecSuiteWillBegin(conf config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (reporter *PRBuilderReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

func (reporter *PRBuilderReporter) SpecWillRun(specSummary *types.SpecSummary) {}
