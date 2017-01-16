package fake

import (
	"bufio"
	"io"
	"sync"

	"github.com/cloudfoundry/dropsonde/log_sender"
	"github.com/cloudfoundry/sonde-go/events"
)

type FakeLogSender struct {
	logs          []Log
	logMessages   []LogMessage
	ReturnError   error
	ReturnChainer log_sender.LogChainer
	sync.RWMutex
}

type Log struct {
	AppId          string
	Message        string
	SourceType     string
	SourceInstance string
	MessageType    string
}

type LogMessage struct {
	Message     []byte
	MessageType events.LogMessage_MessageType
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

func (fls *FakeLogSender) LogMessage(msg []byte, msgType events.LogMessage_MessageType) log_sender.LogChainer {
	fls.Lock()
	defer fls.Unlock()
	fls.logMessages = append(fls.logMessages, LogMessage{
		Message:     msg,
		MessageType: msgType,
	})

	if fls.ReturnChainer != nil {
		c := fls.ReturnChainer
		fls.ReturnChainer = nil
		return c
	}
	return nil
}

func (fls *FakeLogSender) GetLogs() []Log {
	fls.RLock()
	defer fls.RUnlock()

	return fls.logs
}

func (fls *FakeLogSender) GetLogMessages() []LogMessage {
	fls.RLock()
	defer fls.RUnlock()

	return fls.logMessages
}

func (fls *FakeLogSender) Reset() {
	fls.Lock()
	defer fls.Unlock()
	fls.logs = nil
}
