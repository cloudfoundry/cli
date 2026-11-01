package util

import (
	"bytes"
	"sync"
)

// SynchronizedBuffer is a bytes.Buffer that is safe for concurrent use, e.g. when
// a command writes to stdout and stderr from different goroutines into the same
// buffer. The zero value is an empty buffer ready to use.
type SynchronizedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *SynchronizedBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *SynchronizedBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func (sb *SynchronizedBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.buf.Reset()
}
