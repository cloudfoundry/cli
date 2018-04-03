package director

import (
	"fmt"
	"net/http"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TaskClientRequest struct {
	clientRequest         ClientRequest
	taskReporter          TaskReporter
	taskCheckStepDuration time.Duration
}

func NewTaskClientRequest(
	clientRequest ClientRequest,
	taskReporter TaskReporter,
	taskCheckStepDuration time.Duration,
) TaskClientRequest {
	return TaskClientRequest{
		clientRequest:         clientRequest,
		taskReporter:          taskReporter,
		taskCheckStepDuration: taskCheckStepDuration,
	}
}

type taskShortResp struct {
	ID    int    // 165
	State string // e.g. "queued", "processing", "done", "error", "cancelled"
}

func (r taskShortResp) IsRunning() bool {
	return r.State == "queued" || r.State == "processing" || r.State == "cancelling"
}

func (r taskShortResp) IsSuccessfullyDone() bool {
	return r.State == "done"
}

func (r TaskClientRequest) GetResult(path string) (int, []byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Get(path, &taskResp)
	if err != nil {
		return 0, nil, err
	}

	respBody, err := r.waitForResult(taskResp)

	return taskResp.ID, respBody, err
}

func (r TaskClientRequest) PostResult(path string, payload []byte, f func(*http.Request)) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Post(path, payload, f, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) PutResult(path string, payload []byte, f func(*http.Request)) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Put(path, payload, f, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) DeleteResult(path string) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Delete(path, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) WaitForCompletion(id int, type_ string, taskReporter TaskReporter) error {
	taskReporter.TaskStarted(id)

	var taskResp taskShortResp
	var outputOffset int

	defer func() {
		taskReporter.TaskFinished(id, taskResp.State)
	}()

	taskPath := fmt.Sprintf("/tasks/%d", id)

	for {
		err := r.clientRequest.Get(taskPath, &taskResp)
		if err != nil {
			return bosherr.WrapError(err, "Getting task state")
		}

		// retrieve output *after* getting state to make sure
		// it's complete in case of task being finished
		outputOffset, err = r.reportOutputChunk(taskResp.ID, outputOffset, type_, taskReporter)
		if err != nil {
			return bosherr.WrapError(err, "Getting task output")
		}

		if taskResp.IsRunning() {
			time.Sleep(r.taskCheckStepDuration)
			continue
		}

		if taskResp.IsSuccessfullyDone() {
			return nil
		}

		msgFmt := "Expected task '%d' to succeed but state is '%s'"

		return bosherr.Errorf(msgFmt, taskResp.ID, taskResp.State)
	}
}

func (r TaskClientRequest) waitForResult(taskResp taskShortResp) ([]byte, error) {
	err := r.WaitForCompletion(taskResp.ID, "event", r.taskReporter)
	if err != nil {
		return nil, err
	}

	resultPath := fmt.Sprintf("/tasks/%d/output?type=result", taskResp.ID)

	respBody, _, err := r.clientRequest.RawGet(resultPath, nil, nil)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

type taskReporterWriter struct {
	id           int
	totalLen     int
	taskReporter TaskReporter
}

var _ ShouldTrackDownload = &taskReporterWriter{}

func (w *taskReporterWriter) Write(buf []byte) (int, error) {
	bufLen := len(buf)
	if bufLen > 0 {
		w.taskReporter.TaskOutputChunk(w.id, buf)
	}
	w.totalLen += bufLen
	return bufLen, nil
}

func (w taskReporterWriter) ShouldTrackDownload() bool { return false }

func (r TaskClientRequest) reportOutputChunk(id, offset int, type_ string, taskReporter TaskReporter) (int, error) {
	outputPath := fmt.Sprintf("/tasks/%d/output?type=%s", id, type_)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	writer := &taskReporterWriter{id, 0, taskReporter}

	_, resp, err := r.clientRequest.RawGet(outputPath, writer, setHeaders)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return offset, nil
		}

		return 0, err
	}

	return offset + writer.totalLen, nil
}
