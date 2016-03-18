package fake

import (
	"bufio"
	"io"
	"sync"
)

type FakeLogSender struct {
	logs        []Log
	ReturnError error
	sync.RWMutex
}

type Log struct {
	AppId          string
	Message        string
	SourceType     string
	SourceInstance string
	MessageType    string
}

func NewFakeLogSender() *FakeLogSender {
	return &FakeLogSender{}
}

func (fls *FakeLogSender) SendAppLog(appId, message, sourceType, sourceInstance string) error {
	fls.Lock()
	defer fls.Unlock()

	if fls.ReturnError != nil {
		err := fls.ReturnError
		fls.ReturnError = nil

		return err
	}

	fls.logs = append(fls.logs, Log{AppId: appId, Message: message, SourceType: sourceType, SourceInstance: sourceInstance, MessageType: "OUT"})
	return nil
}

func (fls *FakeLogSender) SendAppErrorLog(appId, message, sourceType, sourceInstance string) error {
	fls.Lock()
	defer fls.Unlock()

	if fls.ReturnError != nil {
		err := fls.ReturnError
		fls.ReturnError = nil

		return err
	}

	fls.logs = append(fls.logs, Log{AppId: appId, Message: message, SourceType: sourceType, SourceInstance: sourceInstance, MessageType: "ERR"})
	return nil
}

func (fls *FakeLogSender) ScanLogStream(appId, sourceType, sourceInstance string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		msg := scanner.Text()
		if len(msg) == 0 {
			continue
		}

		fls.Lock()
		fls.logs = append(fls.logs, Log{AppId: appId, SourceType: sourceType, SourceInstance: sourceInstance, MessageType: "OUT", Message: msg})
		fls.Unlock()
	}
}

func (fls *FakeLogSender) ScanErrorLogStream(appId, sourceType, sourceInstance string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		msg := scanner.Text()
		if len(msg) == 0 {
			continue
		}

		fls.Lock()
		fls.logs = append(fls.logs, Log{AppId: appId, SourceType: sourceType, SourceInstance: sourceInstance, MessageType: "ERR", Message: msg})
		fls.Unlock()
	}
}

func (fls *FakeLogSender) GetLogs() []Log {
	fls.RLock()
	defer fls.RUnlock()

	return fls.logs
}

func (fls *FakeLogSender) Reset() {
	fls.Lock()
	defer fls.Unlock()
	fls.logs = nil
}
