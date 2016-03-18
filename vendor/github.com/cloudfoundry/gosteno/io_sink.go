package gosteno

import (
	"bufio"
	"os"
	"sync"
)

type IOSink struct {
	writer *bufio.Writer
	codec  Codec
	file   *os.File

	sync.Mutex
}

func NewIOSink(file *os.File) *IOSink {
	writer := bufio.NewWriter(file)

	x := new(IOSink)
	x.writer = writer
	x.file = file

	return x
}

func NewFileSink(path string) *IOSink {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	return NewIOSink(file)
}

func (x *IOSink) AddRecord(record *Record) {
	bytes, _ := x.codec.EncodeRecord(record)

	x.Lock()
	defer x.Unlock()

	x.writer.Write(bytes)

	// Need to append a newline for IO sink
	x.writer.WriteString("\n")
}

func (x *IOSink) Flush() {
	x.Lock()
	defer x.Unlock()

	x.writer.Flush()
}

func (x *IOSink) SetCodec(codec Codec) {
	x.Lock()
	defer x.Unlock()

	x.codec = codec
}

func (x *IOSink) GetCodec() Codec {
	x.Lock()
	defer x.Unlock()

	return x.codec
}
