package io_helpers

import (
	"bytes"
	"io"
	"os"
	"strings"
)

func SimulateStdin(input string, block func(r io.Reader)) {
	reader, writer := io.Pipe()

	go func() {
		writer.Write([]byte(input))
		defer writer.Close()
	}()

	block(reader)
}

func CaptureOutput(block func()) []string {
	oldSTDOUT := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = oldSTDOUT
	}()

	doneWriting := make(chan bool)
	result := make(chan []string)

	go captureOutputAsyncronously(doneWriting, result, r)

	block()
	w.Close()
	doneWriting <- true
	return <-result
}

/*
 The reason we're doing is that you can't write an infinite amount of bytes into a pipe.
 On some platforms, the limit is fairly high; on other platforms, the limit is infuriatingly small
 (looking at you, Windows). To counteract this, we need to read in a goroutine from one end of
 the pipe and return the result across a channel.
*/
func captureOutputAsyncronously(doneWriting <-chan bool, result chan<- []string, reader io.Reader) {
	var readingString string

	for {
		var buf bytes.Buffer
		io.Copy(&buf, reader)
		readingString += buf.String()

		_, ok := <-doneWriting
		if ok {
			// there is no guarantee that the writer did not
			// write more in between the read above and reading from this channel
			// so we absolute must read once more if we want all the bytes
			var buf bytes.Buffer
			io.Copy(&buf, reader)
			readingString += buf.String()
			break
		}
	}

	result <- strings.Split(readingString, "\n")
}
