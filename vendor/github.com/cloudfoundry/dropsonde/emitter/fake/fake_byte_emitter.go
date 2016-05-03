package fake

import (
	"sync"
)

type FakeByteEmitter struct {
	ReturnError error
	Messages    [][]byte
	mutex       *sync.RWMutex
	isClosed    bool
}

func NewFakeByteEmitter() *FakeByteEmitter {
	return &FakeByteEmitter{mutex: new(sync.RWMutex)}
}
func (f *FakeByteEmitter) Emit(data []byte) (err error) {

	if f.ReturnError != nil {
		err = f.ReturnError
		f.ReturnError = nil
		return
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Messages = append(f.Messages, data)
	return
}

func (f *FakeByteEmitter) GetMessages() (messages [][]byte) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	messages = make([][]byte, len(f.Messages))
	copy(messages, f.Messages)
	return
}

func (f *FakeByteEmitter) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.isClosed = true
}

func (f *FakeByteEmitter) IsClosed() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.isClosed
}
