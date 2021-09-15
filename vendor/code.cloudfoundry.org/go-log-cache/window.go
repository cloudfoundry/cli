package client

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

// Window crawls a reader incrementally to give the Visitor a batch of
// envelopes. Each start time is incremented by the set increment duration if
// that window produced data or not. This is useful when looking for trends
// over time.
func Window(ctx context.Context, v Visitor, w Walker, opts ...WindowOption) {
	c := windowConfig{
		log:      log.New(ioutil.Discard, "", 0),
		width:    time.Hour,
		interval: time.Minute,
	}

	for _, o := range opts {
		o.configure(&c)
	}

	if c.start.IsZero() {
		c.start = time.Now().Add(-c.width)
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		walkCtx, cancel := context.WithTimeout(ctx, c.interval)

		es := w(walkCtx, c.start, c.start.Add(c.width))
		cancel()
		if !v(es) {
			return
		}
		c.start = c.start.Add(c.interval)
	}
}

// Walker walks a reader. It makes several calls to get all the data between
// a boundary of time.
type Walker func(
	ctx context.Context,
	start time.Time,
	end time.Time,
) []*loggregator_v2.Envelope

// BuildWalker captures the sourceID and reader to be used with a Walker.
func BuildWalker(sourceID string, r Reader) Walker {
	return func(ctx context.Context, start, end time.Time) []*loggregator_v2.Envelope {
		var results []*loggregator_v2.Envelope
		Walk(ctx, sourceID, func(e []*loggregator_v2.Envelope) bool {
			results = append(results, e...)
			return true
		}, r,
			WithWalkStartTime(start),
			WithWalkEndTime(end),
		)

		return results
	}
}

// WindowOption configures the Window algorithm.
type WindowOption interface {
	configure(*windowConfig)
}

// WithWindowLogger is used to set the logger for the Walk. It defaults to
// not logging.
func WithWindowLogger(l *log.Logger) WindowOption {
	return windowOptionFunc(func(c *windowConfig) {
		c.log = l
	})
}

// WithWindowStartTime sets the start time of the query. It defaults to
// Now-Width.
func WithWindowStartTime(t time.Time) WindowOption {
	return windowOptionFunc(func(c *windowConfig) {
		c.start = t
	})
}

// WithWindowWidth sets the width (end-start=width) of the query. It defaults
// to an hour.
func WithWindowWidth(w time.Duration) WindowOption {
	return windowOptionFunc(func(c *windowConfig) {
		c.width = w
	})
}

// WithWindowInterval sets the duration to advance the start and end of the
// query. It defaults to a minute.
func WithWindowInterval(i time.Duration) WindowOption {
	return windowOptionFunc(func(c *windowConfig) {
		c.interval = i
	})
}

type windowConfig struct {
	log      *log.Logger
	start    time.Time
	width    time.Duration
	interval time.Duration
}

// windowOptionFunc enables functions to implement WindowOption
type windowOptionFunc func(c *windowConfig)

// configure implements WindowOption.
func (f windowOptionFunc) configure(c *windowConfig) {
	f(c)
}
