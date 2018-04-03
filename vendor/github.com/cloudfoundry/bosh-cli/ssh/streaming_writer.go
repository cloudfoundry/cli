package ssh

import (
	"fmt"
	"io"

	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type StreamingWriter struct {
	comboWriter *boshui.ComboWriter
}

func NewStreamingWriter(comboWriter *boshui.ComboWriter) *StreamingWriter {
	return &StreamingWriter{comboWriter: comboWriter}
}

func (w StreamingWriter) ForInstance(jobName, indexOrID string) InstanceWriter {
	return streamingInstanceWriter{jobName: jobName, indexOrID: indexOrID, comboWriter: w.comboWriter}
}

func (w StreamingWriter) Flush() {}

type streamingInstanceWriter struct {
	jobName   string
	indexOrID string

	comboWriter *boshui.ComboWriter
}

func (w streamingInstanceWriter) Stdout() io.Writer {
	return w.comboWriter.Writer(fmt.Sprintf("%s/%s: stdout | ", w.jobName, w.indexOrID))
}

func (w streamingInstanceWriter) Stderr() io.Writer {
	return w.comboWriter.Writer(fmt.Sprintf("%s/%s: stderr | ", w.jobName, w.indexOrID))
}

func (w streamingInstanceWriter) End(exitStatus int, err error) {}
