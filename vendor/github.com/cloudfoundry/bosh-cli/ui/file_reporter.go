package ui

import (
	"io"

	"github.com/cheggaaa/pb"
)

type FileReporter struct {
	ui UI
}

func NewFileReporter(ui UI) FileReporter {
	return FileReporter{ui: ui}
}

type ReadSeekCloser interface {
	io.Seeker
	io.ReadCloser
}

func (r FileReporter) Write(b []byte) (int, error) {
	r.ui.BeginLinef("%s", b)
	return len(b), nil
}

func (r FileReporter) TrackUpload(size int64, reader io.ReadCloser) ReadSeekCloser {
	return &ReadCloserProxy{reader: reader, bar: r.buildBar(size), ui: r.ui}
}

func (r FileReporter) TrackDownload(size int64, writer io.Writer) io.Writer {
	return io.MultiWriter(writer, r.buildBar(size))
}

func (r FileReporter) buildBar(size int64) *pb.ProgressBar {
	bar := pb.New(int(size))
	bar.Output = r
	bar.NotPrint = true
	bar.ShowCounters = false
	bar.ShowTimeLeft = true
	bar.ShowSpeed = true
	bar.SetWidth(80)
	bar.SetMaxWidth(80)
	bar.SetUnits(pb.U_BYTES)
	bar.Format("\x00#\x00#\x00 \x00")
	bar.Start()
	return bar
}

type ReadCloserProxy struct {
	reader io.ReadCloser
	bar    *pb.ProgressBar
	ui     UI
}

func (p ReadCloserProxy) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := p.reader.(io.Seeker)
	if ok {
		return seeker.Seek(offset, whence)
	}

	return 0, nil
}

func (p *ReadCloserProxy) Read(bs []byte) (int, error) {
	n, err := p.reader.Read(bs)
	p.bar.Add(n)
	return n, err
}

func (p *ReadCloserProxy) Close() error {
	err := p.reader.Close()
	p.bar.Finish()
	p.ui.BeginLinef("\n")
	return err
}
