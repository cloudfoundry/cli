package metric_sender

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

const (
	maxTagLen = 256
	maxTags   = 10
)

type EventEmitter interface {
	Emit(events.Event) error
	EmitEnvelope(*events.Envelope) error
	Origin() string
}

type ValueChainer interface {
	SetTag(key, value string) ValueChainer
	Send() error
}

type ContainerMetricChainer interface {
	SetTag(key, value string) ContainerMetricChainer
	Send() error
}

type CounterChainer interface {
	SetTag(key, value string) CounterChainer
	Increment() error
	Add(delta uint64) error
}

// A MetricSender emits metric events.
type MetricSender struct {
	eventEmitter EventEmitter
}

// NewMetricSender instantiates a MetricSender with the given EventEmitter.
func NewMetricSender(eventEmitter EventEmitter) *MetricSender {
	return &MetricSender{eventEmitter: eventEmitter}
}

// Send sends an events.Event.
func (ms *MetricSender) Send(ev events.Event) error {
	return ms.eventEmitter.Emit(ev)
}

// SendValue sends a metric with the given name, value and unit. See
// http://metrics20.org/spec/#units for a specification of acceptable units.
// Returns an error if one occurs while sending the event.
func (ms *MetricSender) SendValue(name string, value float64, unit string) error {
	return ms.eventEmitter.Emit(&events.ValueMetric{Name: &name, Value: &value, Unit: &unit})
}

// IncrementCounter sends an event to increment the named counter by one.
// Maintaining the value of the counter is the responsibility of the receiver of
// the event, not the process that includes this package.
func (ms *MetricSender) IncrementCounter(name string) error {
	return ms.AddToCounter(name, 1)
}

// AddToCounter sends an event to increment the named counter by the specified
// (positive) delta. Maintaining the value of the counter is the responsibility
// of the receiver, as with IncrementCounter.
func (ms *MetricSender) AddToCounter(name string, delta uint64) error {
	return ms.eventEmitter.Emit(&events.CounterEvent{Name: &name, Delta: &delta})
}

// SendContainerMetric sends a metric that records resource usage of an app in a container.
// The container is identified by the applicationId and the instanceIndex. The resource
// metrics are CPU percentage, memory and disk usage in bytes. Returns an error if one occurs
// when sending the metric.
func (ms *MetricSender) SendContainerMetric(applicationId string, instanceIndex int32, cpuPercentage float64, memoryBytes uint64, diskBytes uint64) error {
	return ms.eventEmitter.Emit(&events.ContainerMetric{ApplicationId: &applicationId, InstanceIndex: &instanceIndex, CpuPercentage: &cpuPercentage, MemoryBytes: &memoryBytes, DiskBytes: &diskBytes})
}

// Value creates a value metric that can be manipulated via cascading calls
// and then sent.
func (ms *MetricSender) Value(name string, value float64, unit string) ValueChainer {
	chainer := valueChainer{}
	chainer.emitter = ms.eventEmitter
	chainer.envelope = &events.Envelope{
		Origin:    proto.String(ms.eventEmitter.Origin()),
		EventType: events.Envelope_ValueMetric.Enum(),
		ValueMetric: &events.ValueMetric{
			Name:  proto.String(name),
			Value: proto.Float64(value),
			Unit:  proto.String(unit),
		},
	}
	return chainer
}

// ContainerMetric creates a container metric that can be manipulated via
// cascading calls and then sent.
func (ms *MetricSender) ContainerMetric(appID string, instance int32, cpu float64, mem, disk uint64) ContainerMetricChainer {
	chainer := containerMetricChainer{}
	chainer.emitter = ms.eventEmitter
	chainer.envelope = &events.Envelope{
		Origin:    proto.String(ms.eventEmitter.Origin()),
		EventType: events.Envelope_ContainerMetric.Enum(),
		ContainerMetric: &events.ContainerMetric{
			ApplicationId: proto.String(appID),
			InstanceIndex: proto.Int32(instance),
			CpuPercentage: proto.Float64(cpu),
			MemoryBytes:   proto.Uint64(mem),
			DiskBytes:     proto.Uint64(disk),
		},
	}
	return chainer
}

// Counter creates a counter event that can be manipulated via cascading calls
// and then sent via Increment or Add.
func (ms *MetricSender) Counter(name string) CounterChainer {
	chainer := counterChainer{}
	chainer.emitter = ms.eventEmitter
	chainer.envelope = &events.Envelope{
		Origin:    proto.String(ms.eventEmitter.Origin()),
		EventType: events.Envelope_CounterEvent.Enum(),
		CounterEvent: &events.CounterEvent{
			Name: proto.String(name),
		},
	}
	return chainer
}

type envelopeEmitter interface {
	EmitEnvelope(*events.Envelope) error
}

type chainer struct {
	emitter  envelopeEmitter
	envelope *events.Envelope
	err      error
}

func (c chainer) SetTag(key, value string) chainer {
	if c.envelope.Tags == nil {
		c.envelope.Tags = make(map[string]string)
	}
	if utf8.RuneCountInString(key) > maxTagLen || utf8.RuneCountInString(value) > maxTagLen {
		return chainer{
			err: fmt.Errorf("Tag exceeds max length of %d", maxTagLen),
		}
	}

	c.envelope.Tags[key] = value
	if len(c.envelope.Tags) > maxTags {
		return chainer{
			err: fmt.Errorf("Too many tags. Max of %d", maxTags),
		}
	}
	return c
}

func (c chainer) Send() error {
	if c.err != nil {
		return c.err
	}

	c.envelope.Timestamp = proto.Int64(time.Now().UnixNano())
	return c.emitter.EmitEnvelope(c.envelope)
}

type valueChainer struct {
	chainer
}

func (c valueChainer) SetTag(key, value string) ValueChainer {
	c.chainer = c.chainer.SetTag(key, value)
	return c
}

type containerMetricChainer struct {
	chainer
}

func (c containerMetricChainer) SetTag(key, value string) ContainerMetricChainer {
	c.chainer = c.chainer.SetTag(key, value)
	return c
}

type counterChainer struct {
	chainer
}

func (c counterChainer) SetTag(key, value string) CounterChainer {
	c.chainer = c.chainer.SetTag(key, value)
	return c
}

func (c counterChainer) Add(delta uint64) error {
	if c.err != nil {
		return c.err
	}

	c.envelope.CounterEvent.Delta = proto.Uint64(delta)
	return c.chainer.Send()
}

func (c counterChainer) Increment() error {
	if c.err != nil {
		return c.err
	}

	return c.Add(1)
}
