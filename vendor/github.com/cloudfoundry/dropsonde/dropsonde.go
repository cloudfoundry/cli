// Package dropsonde provides sensible defaults for using dropsonde.
//
// The default HTTP transport is instrumented, as well as some basic stats about
// the Go runtime. The default emitter sends events over UDP.
//
// Use
//
// dropsonde.Initialize("localhost:3457", origins...)
//
// to initialize. See package metrics and logs for other usage.
package dropsonde

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/envelope_sender"
	"github.com/cloudfoundry/dropsonde/envelopes"
	"github.com/cloudfoundry/dropsonde/instrumented_handler"
	"github.com/cloudfoundry/dropsonde/instrumented_round_tripper"
	"github.com/cloudfoundry/dropsonde/log_sender"
	"github.com/cloudfoundry/dropsonde/logs"
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"github.com/cloudfoundry/dropsonde/metricbatcher"
	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/dropsonde/runtime_stats"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
)

var autowiredEmitter emitter.EventEmitter

const (
	statsInterval        = 10 * time.Second
	defaultBatchInterval = 5 * time.Second
	originDelimiter      = "/"
)

func init() {
	autowiredEmitter = &NullEventEmitter{}
}

// Initialize creates default emitters and instruments the default HTTP
// transport.
//
// The origin variable is required and specifies the
// source name for all metrics emitted by this process. If it is not set, the
// program will run normally but will not emit metrics.
//
// The destination variable sets the host and port to
// which metrics are sent. It is optional, and defaults to DefaultDestination.
func Initialize(destination string, origin ...string) error {
	emitter, err := createDefaultEmitter(strings.Join(origin, originDelimiter), destination)
	if err != nil {
		autowiredEmitter = &NullEventEmitter{}
		return err
	}

	autowiredEmitter = emitter
	initialize()

	return nil
}

// InitializeWithEmitter sets up Dropsonde with the passed emitter, instead of
// creating one.
func InitializeWithEmitter(emitter emitter.EventEmitter) {
	autowiredEmitter = emitter
	initialize()
}

// AutowiredEmitter exposes the emitter used by Dropsonde after its initialization.
func AutowiredEmitter() emitter.EventEmitter {
	return autowiredEmitter
}

// InstrumentedHandler returns a Handler pre-configured to emit HTTP server
// request metrics to AutowiredEmitter.
func InstrumentedHandler(handler http.Handler) http.Handler {
	return instrumented_handler.InstrumentedHandler(handler, autowiredEmitter)
}

// InstrumentedRoundTripper returns a RoundTripper pre-configured to emit
// HTTP client request metrics to AutowiredEmitter.
func InstrumentedRoundTripper(roundTripper http.RoundTripper) http.RoundTripper {
	return instrumented_round_tripper.InstrumentedRoundTripper(roundTripper, autowiredEmitter)
}

func initialize() {
	emitter := AutowiredEmitter()
	sender := metric_sender.NewMetricSender(emitter)
	batcher := metricbatcher.New(sender, defaultBatchInterval)
	metrics.Initialize(sender, batcher)
	logs.Initialize(log_sender.NewLogSender(AutowiredEmitter(), gosteno.NewLogger("dropsonde/logs")))
	envelopes.Initialize(envelope_sender.NewEnvelopeSender(emitter))
	go runtime_stats.NewRuntimeStats(autowiredEmitter, statsInterval).Run(nil)
	http.DefaultTransport = InstrumentedRoundTripper(http.DefaultTransport)
}

func createDefaultEmitter(origin, destination string) (emitter.EventEmitter, error) {
	if len(origin) == 0 {
		return nil, errors.New("Failed to initialize dropsonde: origin variable not set")
	}

	if len(destination) == 0 {
		return nil, errors.New("Failed to initialize dropsonde: destination variable not set")
	}

	udpEmitter, err := emitter.NewUdpEmitter(destination)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize dropsonde: %v", err.Error())
	}

	return emitter.NewEventEmitter(udpEmitter, origin), nil
}

// NullEventEmitter is used when no event emission is desired. See
// http://en.wikipedia.org/wiki/Null_Object_pattern.
type NullEventEmitter struct{}

// Emit is called to send an event to a remote host. On NullEventEmitter,
// it is a no-op.
func (*NullEventEmitter) Emit(events.Event) error {
	return nil
}

// EmitEnvelope is called to send an envelope to a remote host. On NullEventEmitter,
// it is a no-op.
func (*NullEventEmitter) EmitEnvelope(*events.Envelope) error {
	return nil
}

// Close ceases emitter operations. On NullEventEmitter, it is a no-op.
func (*NullEventEmitter) Close() {}
