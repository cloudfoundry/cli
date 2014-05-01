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
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldSTDOUT
	}()

	block()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return strings.Split(buf.String(), "\n")
}
