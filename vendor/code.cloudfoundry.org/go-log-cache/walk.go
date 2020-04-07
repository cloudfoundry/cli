package client

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
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
	c := &WalkConfig{
		Log:     log.New(ioutil.Discard, "", 0),
		Backoff: AlwaysDoneBackoff{},
		Delay:   time.Second,
	}

	for _, o := range opts {
		o(c)
	}

	var readOpts []ReadOption
	if !c.End.IsZero() {
		readOpts = append(readOpts, WithEndTime(c.End))
	}

	if c.Limit != nil {
		readOpts = append(readOpts, WithLimit(*c.Limit))
	}

	if c.EnvelopeTypes != nil {
		readOpts = append(readOpts, WithEnvelopeTypes(c.EnvelopeTypes...))
	}

	if c.NameFilter != "" {
		readOpts = append(readOpts, WithNameFilter(c.NameFilter))
	}

	var receivedEmpty bool

	for {
		es, err := r(ctx, sourceID, time.Unix(0, c.Start), readOpts...)
		if err != nil && ctx.Err() != nil {
			// Context cancelled
			return
		}

		if err != nil {
			c.Log.Print(err)
			if !c.Backoff.OnErr(err) {
				return
			}
			continue
		}

		if c.End.IsZero() || !receivedEmpty {
			// Prune envelopes for any that are too new or from the future.
			withDelay := time.Now().Add(-c.Delay).UnixNano()
			for i := len(es) - 1; i >= 0; i-- {
				if es[i].GetTimestamp() <= withDelay {
					// The rest of the envelopes aren't too new.
					break
				}

				// Envelope is too new. Throw it away.
				es = es[:i]
			}
		}

		if !c.End.IsZero() {
			for i := len(es) - 1; i >= 0; i-- {
				if es[i].GetTimestamp() < c.End.UnixNano() {
					break
				}

				es = es[:i]
			}
		}

		if len(es) == 0 {
			receivedEmpty = true
			if !c.Backoff.OnEmpty() {
				return
			}
			continue
		}

		c.Backoff.Reset()
		receivedEmpty = false

		// If visitor is done or the next timestamp would be outside of our
		// window (only when End is set), then be done.
		if !v(es) || (!c.End.IsZero() && es[len(es)-1].Timestamp+1 >= c.End.UnixNano()) {
			return
		}

		c.Start = es[len(es)-1].Timestamp + 1
	}
}

// WalkOption overrides defaults for Walk.
type WalkOption func(config *WalkConfig)

// WithWalkLogger is used to set the logger for the Walk. It defaults to
// not logging.
func WithWalkLogger(l *log.Logger) WalkOption {
	return func(c *WalkConfig) {
		c.Log = l
	}
}

// WithWalkStartTime sets the Start time of the query.
func WithWalkStartTime(t time.Time) WalkOption {
	return func(c *WalkConfig) {
		c.Start = t.UnixNano()
	}
}

// WithWalkEndTime sets the End time of the query. Once reached, Walk will
// exit.
func WithWalkEndTime(t time.Time) WalkOption {
	return func(c *WalkConfig) {
		c.End = t
	}
}

// WithWalkLimit sets the Limit of the query.
func WithWalkLimit(limit int) WalkOption {
	return func(c *WalkConfig) {
		c.Limit = &limit
	}
}

// WithWalkEnvelopeType sets the envelope_types of the query.
func WithWalkEnvelopeTypes(t ...logcache_v1.EnvelopeType) WalkOption {
	return func(c *WalkConfig) {
		c.EnvelopeTypes = t
	}
}

func WithWalkNameFilter(nameFilter string) WalkOption {
	return func(c *WalkConfig) {
		c.NameFilter = nameFilter
	}
}

// WithWalkBackoff sets the Backoff strategy for an empty batch or error. It
// defaults to stopping on an error or empty batch via AlwaysDoneBackoff.
func WithWalkBackoff(b Backoff) WalkOption {
	return func(c *WalkConfig) {
		c.Backoff = b
	}
}

// WithWalkDelay sets the value that the walk algorithm will consider "old"
// enough. If an envelope has a timestamp that has a value that is greater
// than time.Now().Add(-Delay), it will be considered too "new", and not
// included. This protects from distributed clocks causing data to be skipped.
// Defaults to 1 second.
func WithWalkDelay(delay time.Duration) WalkOption {
	return func(c *WalkConfig) {
		c.Delay = delay
	}
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

type WalkConfig struct {
	Log           *log.Logger
	Backoff       Backoff
	Start         int64
	End           time.Time
	Limit         *int
	EnvelopeTypes []logcache_v1.EnvelopeType
	Delay         time.Duration
	NameFilter    string
}
