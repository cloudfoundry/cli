package test_helpers

type FakeDebugPrinter struct {
	Messages []*fakeDebugPrinterMessage
}

type fakeDebugPrinterMessage struct {
	Title, Body string
}

func (p *FakeDebugPrinter) Print(title, body string) {
	message := &fakeDebugPrinterMessage{title, body}
	p.Messages = append(p.Messages, message)
}
