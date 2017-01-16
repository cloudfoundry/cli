package instrumented_handler

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
)

type EventEmitter interface {
	Emit(events.Event) error
}

type instrumentedHandler struct {
	handler http.Handler
	emitter EventEmitter
}

// InstrumentedHandler is a helper for creating an instrumented http.Handler
// which will delegate to the given http.Handler.
func InstrumentedHandler(handler http.Handler, emitter EventEmitter) http.Handler {
	return &instrumentedHandler{handler, emitter}
}

// ServeHTTP wraps the given http.Handler ServerHTTP function.  It provides
// accounting metrics for the http.Request / http.Response life-cycle
func (ih *instrumentedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestId, err := uuid.ParseHex(req.Header.Get("X-Vcap-Request-Id"))
	if err != nil {
		requestId, err = GenerateUuid()
		if err != nil {
			log.Printf("failed to generate request ID: %v\n", err)
			requestId = &uuid.UUID{}
		}
		req.Header.Set("X-Vcap-Request-Id", requestId.String())
	}
	rw.Header().Set("X-Vcap-Request-Id", requestId.String())

	startTime := time.Now()

	instrumentedWriter := &instrumentedResponseWriter{writer: rw, statusCode: 200}
	ih.handler.ServeHTTP(instrumentedWriter, req)

	startStopEvent := factories.NewHttpStartStop(req, instrumentedWriter.statusCode, instrumentedWriter.contentLength, events.PeerType_Server, requestId)
	startStopEvent.StartTimestamp = proto.Int64(startTime.UnixNano())

	err = ih.emitter.Emit(startStopEvent)
	if err != nil {
		log.Printf("failed to emit startstop event: %v\n", err)
	}
}

type instrumentedResponseWriter struct {
	writer        http.ResponseWriter
	contentLength int64
	statusCode    int
}

func (irw *instrumentedResponseWriter) Header() http.Header {
	return irw.writer.Header()
}

func (irw *instrumentedResponseWriter) Write(data []byte) (int, error) {
	writeCount, err := irw.writer.Write(data)
	irw.contentLength += int64(writeCount)
	return writeCount, err
}

func (irw *instrumentedResponseWriter) WriteHeader(statusCode int) {
	irw.statusCode = statusCode
	irw.writer.WriteHeader(statusCode)
}

func (irw *instrumentedResponseWriter) Flush() {
	flusher, ok := irw.writer.(http.Flusher)

	if !ok {
		panic("Called Flush on an InstrumentedResponseWriter that wraps a non-Flushable writer.")
	}

	flusher.Flush()
}

func (irw *instrumentedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := irw.writer.(http.Hijacker)

	if !ok {
		panic("Called Hijack on an InstrumentedResponseWriter that wraps a non-Hijackable writer")
	}

	return hijacker.Hijack()
}

func (irw *instrumentedResponseWriter) CloseNotify() <-chan bool {
	notifier, ok := irw.writer.(http.CloseNotifier)

	if !ok {
		panic("Called CloseNotify on an InstrumentedResponseWriter that wraps a non-CloseNotifiable writer")
	}

	return notifier.CloseNotify()
}

var GenerateUuid = uuid.NewV4
