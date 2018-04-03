package ssh

import (
	"bytes"
	"fmt"
	"io"

	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ResultsWriter struct {
	ui boshui.UI

	instances []*resultsInstanceWriter
}

func NewResultsWriter(ui boshui.UI) *ResultsWriter {
	return &ResultsWriter{ui: ui}
}

func (w *ResultsWriter) ForInstance(jobName, indexOrID string) InstanceWriter {
	w.instances = append(w.instances, newBufferedInstanceWriter(jobName, indexOrID))
	return w.instances[len(w.instances)-1]
}

func (w *ResultsWriter) Flush() {
	table := boshtbl.Table{
		Content: "results",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Instance"),
			boshtbl.NewHeader("Stdout"),
			boshtbl.NewHeader("Stderr"),
			boshtbl.NewHeader("Exit Code"),
			boshtbl.NewHeader("Error"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},

		Transpose: true,
	}

	for _, inst := range w.instances {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(inst.Instance()),
			boshtbl.NewValueString(inst.StdoutAsString()),
			boshtbl.NewValueString(inst.StderrAsString()),
			boshtbl.NewValueInt(inst.ExitStatus()),
			boshtbl.NewValueError(inst.Error()),
		})
	}

	w.ui.PrintTable(table)
}

type resultsInstanceWriter struct {
	jobName   string
	indexOrID string

	stdout *bytes.Buffer
	stderr *bytes.Buffer

	exitStatus int
	error      error
}

func newBufferedInstanceWriter(jobName, indexOrID string) *resultsInstanceWriter {
	return &resultsInstanceWriter{
		jobName:   jobName,
		indexOrID: indexOrID,

		stdout: bytes.NewBufferString(""),
		stderr: bytes.NewBufferString(""),
	}
}

func (w *resultsInstanceWriter) Instance() string {
	return fmt.Sprintf("%s/%s", w.jobName, w.indexOrID)
}

func (w *resultsInstanceWriter) Stdout() io.Writer      { return w.stdout }
func (w *resultsInstanceWriter) StdoutAsString() string { return w.stdout.String() }

func (w *resultsInstanceWriter) Stderr() io.Writer      { return w.stderr }
func (w *resultsInstanceWriter) StderrAsString() string { return w.stderr.String() }

func (w *resultsInstanceWriter) End(exitStatus int, err error) {
	w.exitStatus = exitStatus
	w.error = err
}

func (w *resultsInstanceWriter) ExitStatus() int { return w.exitStatus }
func (w *resultsInstanceWriter) Error() error    { return w.error }
