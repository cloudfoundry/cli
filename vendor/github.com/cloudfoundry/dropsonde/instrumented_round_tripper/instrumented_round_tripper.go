package instrumented_round_tripper

import (
	"log"
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

type instrumentedRoundTripper struct {
	roundTripper http.RoundTripper
	emitter      EventEmitter
}

type instrumentedCancelableRoundTripper struct {
	instrumentedRoundTripper *instrumentedRoundTripper
}

// InstrumentedRoundTripper is a helper for creating a "net/http".RoundTripper
// which will delegate to the given RoundTripper.
func InstrumentedRoundTripper(roundTripper http.RoundTripper, emitter EventEmitter) http.RoundTripper {
	irt := &instrumentedRoundTripper{roundTripper, emitter}

	_, ok := roundTripper.(canceler)
	if ok {
		return &instrumentedCancelableRoundTripper{
			instrumentedRoundTripper: irt,
		}
	}

	return irt
}

// RoundTrip wraps the RoundTrip function of the given RoundTripper.  It
// provides accounting metrics for the http.Request / http.Response life-cycle.
// Callers of RoundTrip are responsible for setting the ‘X-Vcap-Request-Id’
// field in the request header if they have one.  Callers are also responsible
// for setting the ‘X-CF-ApplicationID’ and ‘X-CF-InstanceIndex’ fields in the
// request header if they are known.
func (irt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()
	requestId := req.Header.Get("X-Vcap-Request-Id")
	if requestId == "" {
		requestIdGuid, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		requestId = requestIdGuid.String()
	}

	req.Header.Set("X-Vcap-Request-Id", requestId)

	resp, roundTripErr := irt.roundTripper.RoundTrip(req)

	var statusCode int
	var contentLength int64
	if roundTripErr == nil {
		statusCode = resp.StatusCode
		contentLength = resp.ContentLength
	}

	id, err := uuid.ParseHex(requestId)
	if err != nil {
		return nil, err
	}

	httpStartStop := factories.NewHttpStartStop(req, statusCode, contentLength, events.PeerType_Client, id)
	httpStartStop.StartTimestamp = proto.Int64(startTime.UnixNano())

	err = irt.emitter.Emit(httpStartStop)
	if err != nil {
		log.Printf("failed to emit startstop event: %v\n", err)
	}

	return resp, roundTripErr
}

func (icrt *instrumentedCancelableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return icrt.instrumentedRoundTripper.RoundTrip(req)
}

func (icrt *instrumentedCancelableRoundTripper) CancelRequest(req *http.Request) {
	cancelableTransport := icrt.instrumentedRoundTripper.roundTripper.(canceler)
	cancelableTransport.CancelRequest(req)
}

type canceler interface {
	CancelRequest(*http.Request)
}
