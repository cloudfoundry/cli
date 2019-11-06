package client

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

// Reader reads envelopes from LogCache. It will be invoked by Walker several
// time to traverse the length of the cache.
type Reader func(
	ctx context.Context,
	sourceID string,
	start time.Time,
	opts ...ReadOption,
) ([]*loggregator_v2.Envelope, error)

// Visitor is invoked for each envelope batch. If the function returns false,
// it doesn't make any more requests. Otherwise it reaches out for the next
// batch of envelopes.
type Visitor func([]*loggregator_v2.Envelope) bool

// Walk reads from the LogCache until the Visitor returns false.
func Walk(ctx context.Context, sourceID string, v Visitor, r Reader, opts ...WalkOption) {
	c := &walkConfig{
		log:     log.New(ioutil.Discard, "", 0),
		backoff: AlwaysDoneBackoff{},
		delay:   time.Second,
	}

	for _, o := range opts {
		o.configure(c)
	}

	var readOpts []ReadOption
	if !c.end.IsZero() {
		readOpts = append(readOpts, WithEndTime(c.end))
	}

	if c.limit != nil {
		readOpts = append(readOpts, WithLimit(*c.limit))
	}

	if c.envelopeTypes != nil {
		readOpts = append(readOpts, WithEnvelopeTypes(c.envelopeTypes...))
	}

	if c.nameFilter != "" {
		readOpts = append(readOpts, WithNameFilter(c.nameFilter))
	}

	var receivedEmpty bool

	for {
		es, err := r(ctx, sourceID, time.Unix(0, c.start), readOpts...)
		if err != nil && ctx.Err() != nil {
			// Context cancelled
			return
		}

		if err != nil {
			c.log.Print(err)
			if !c.backoff.OnErr(err) {
				return
			}
			continue
		}

		if c.end.IsZero() || !receivedEmpty {
			// Prune envelopes for any that are too new or from the future.
			withDelay := time.Now().Add(-c.delay).UnixNano()
			for i := len(es) - 1; i >= 0; i-- {
				if es[i].GetTimestamp() <= withDelay {
					// The rest of the envelopes aren't too new.
					break
				}

				// Envelope is too new. Throw it away.
				es = es[:i]
			}
		}

		if !c.end.IsZero() {
			for i := len(es) - 1; i >= 0; i-- {
				if es[i].GetTimestamp() < c.end.UnixNano() {
					break
				}

				es = es[:i]
			}
		}

		if len(es) == 0 {
			receivedEmpty = true
			if !c.backoff.OnEmpty() {
				return
			}
			continue
		}

		c.backoff.Reset()
		receivedEmpty = false

		// If visitor is done or the next timestamp would be outside of our
		// window (only when end is set), then be done.
		if !v(es) || (!c.end.IsZero() && es[len(es)-1].Timestamp+1 >= c.end.UnixNano()) {
			return
		}

		c.start = es[len(es)-1].Timestamp + 1
	}
}

// WalkOption overrides defaults for Walk.
type WalkOption interface {
	configure(*walkConfig)
}

// WithWalkLogger is used to set the logger for the Walk. It defaults to
// not logging.
func WithWalkLogger(l *log.Logger) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.log = l
	})
}

// WithWalkStartTime sets the start time of the query.
func WithWalkStartTime(t time.Time) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.start = t.UnixNano()
	})
}

// WithWalkEndTime sets the end time of the query. Once reached, Walk will
// exit.
func WithWalkEndTime(t time.Time) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.end = t
	})
}

// WithWalkLimit sets the limit of the query.
func WithWalkLimit(limit int) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.limit = &limit
	})
}

// WithWalkEnvelopeType sets the envelope_types of the query.
func WithWalkEnvelopeTypes(t ...logcache_v1.EnvelopeType) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.envelopeTypes = t
	})
}

func WithWalkNameFilter(nameFilter string) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.nameFilter = nameFilter
	})
}

// WithWalkBackoff sets the backoff strategy for an empty batch or error. It
// defaults to stopping on an error or empty batch via AlwaysDoneBackoff.
func WithWalkBackoff(b Backoff) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.backoff = b
	})
}

// WithWalkDelay sets the value that the walk algorithm will consider "old"
// enough. If an envelope has a timestamp that has a value that is greater
// than time.Now().Add(-delay), it will be considered too "new", and not
// included. This protects from distributed clocks causing data to be skipped.
// Defaults to 1 second.
func WithWalkDelay(delay time.Duration) WalkOption {
	return walkOptionFunc(func(c *walkConfig) {
		c.delay = delay
	})
}

// Backoff is used to determine what to do if there is an empty batch or
// error. If there is an error, it will be passed to the method OnErr. If there is
// not an error and just an empty batch, the method OnEmpty will be invoked. If
// the either method returns false, then Walk exits. On a successful read that
// has envelopes, Reset will be invoked.
type Backoff interface {
	OnErr(error) bool
	OnEmpty() bool
	Reset()
}

// AlwaysDoneBackoff returns false for both OnErr and OnEmpty.
type AlwaysDoneBackoff struct{}

// OnErr implements Backoff.
func (b AlwaysDoneBackoff) OnErr(error) bool {
	return false
}

// OnEmpty implements Backoff.
func (b AlwaysDoneBackoff) OnEmpty() bool {
	return false
}

// Reset implements Backoff.
func (b AlwaysDoneBackoff) Reset() {}

// AlwaysRetryBackoff returns true for both OnErr and OnEmpty after sleeping
// the given interval.
type AlwaysRetryBackoff struct {
	interval time.Duration
}

// NewAlwaysRetryBackoff returns a new AlwaysRetryBackoff.
func NewAlwaysRetryBackoff(interval time.Duration) AlwaysRetryBackoff {
	return AlwaysRetryBackoff{
		interval: interval,
	}
}

// OnErr implements Backoff.
func (b AlwaysRetryBackoff) OnErr(error) bool {
	time.Sleep(b.interval)
	return true
}

// OnEmpty implements Backoff.
func (b AlwaysRetryBackoff) OnEmpty() bool {
	time.Sleep(b.interval)
	return true
}

// Reset implements Backoff.
func (b AlwaysRetryBackoff) Reset() {}

// RetryBackoff returns true for both OnErr and OnEmpty after sleeping
// the given interval for a limited number of times.
type RetryBackoff struct {
	interval time.Duration
	maxCount int
	count    int
	onlyErr  bool
}

// NewRetryBackoff returns a new RetryBackoff.
func NewRetryBackoff(interval time.Duration, maxCount int) *RetryBackoff {
	return &RetryBackoff{
		interval: interval,
		maxCount: maxCount,
	}
}

// NewRetryBackoffOnErr returns a new RetryBackoff that only backs off no
// errors.
func NewRetryBackoffOnErr(interval time.Duration, maxCount int) *RetryBackoff {
	return &RetryBackoff{
		interval: interval,
		maxCount: maxCount,
		onlyErr:  true,
	}
}

// OnErr implements Backoff.
func (b *RetryBackoff) OnErr(error) bool {
	b.count++
	if b.count >= b.maxCount {
		return false
	}

	time.Sleep(b.interval)
	return true
}

// OnEmpty implements Backoff.
func (b *RetryBackoff) OnEmpty() bool {
	if b.onlyErr {
		return false
	}

	b.count++
	if b.count >= b.maxCount {
		return false
	}

	time.Sleep(b.interval)
	return true
}

// Reset implements Backoff.
func (b *RetryBackoff) Reset() {
	b.count = 0
}

type walkOptionFunc func(*walkConfig)

func (f walkOptionFunc) configure(c *walkConfig) {
	f(c)
}

type walkConfig struct {
	log           *log.Logger
	backoff       Backoff
	start         int64
	end           time.Time
	limit         *int
	envelopeTypes []logcache_v1.EnvelopeType
	delay         time.Duration
	nameFilter    string
}
