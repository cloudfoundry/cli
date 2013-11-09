package trace_test

import (
	"bytes"
	"cf/trace"
	"fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestTraceSetToFalse(t *testing.T) {
	stdOut := bytes.NewBuffer([]byte{})
	trace.SetStdout(stdOut)

	os.Setenv(trace.CF_TRACE, "false")

	logger := trace.NewLogger()
	logger.Print("hello world")

	result, _ := ioutil.ReadAll(stdOut)
	assert.Equal(t, string(result), "")
}

func TestTraceSetToTrue(t *testing.T) {
	stdOut := bytes.NewBuffer([]byte{})
	trace.SetStdout(stdOut)

	os.Setenv(trace.CF_TRACE, "true")

	logger := trace.NewLogger()
	logger.Print("hello world")

	result, _ := ioutil.ReadAll(stdOut)
	assert.Contains(t, string(result), "hello world")
}

func TestTraceSetToFile(t *testing.T) {
	stdOut := bytes.NewBuffer([]byte{})
	trace.SetStdout(stdOut)

	fileutils.TempFile("trace_test", func(file *os.File, err error) {
		assert.NoError(t, err)
		file.Write([]byte("pre-existing content"))

		os.Setenv(trace.CF_TRACE, file.Name())

		logger := trace.NewLogger()
		logger.Print("hello world")

		file.Seek(0, os.SEEK_SET)
		result, err := ioutil.ReadAll(file)
		assert.NoError(t, err)

		byteString := string(result)
		assert.Contains(t, byteString, "pre-existing content")
		assert.Contains(t, byteString, "hello world")

		result, _ = ioutil.ReadAll(stdOut)
		assert.Equal(t, string(result), "")
	})
}

func TestTraceSetToInvalidFile(t *testing.T) {
	stdOut := bytes.NewBuffer([]byte{})
	trace.SetStdout(stdOut)

	fileutils.TempFile("trace_test", func(file *os.File, err error) {
		assert.NoError(t, err)

		file.Chmod(0000)

		os.Setenv(trace.CF_TRACE, file.Name())

		logger := trace.NewLogger()
		logger.Print("hello world")

		result, _ := ioutil.ReadAll(file)
		assert.Equal(t, string(result), "")

		result, _ = ioutil.ReadAll(stdOut)
		assert.Contains(t, string(result), "hello world")
	})
}
