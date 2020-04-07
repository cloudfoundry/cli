package sharedaction

import (
	"context"
	"time"

	logcache "code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

//go:generate counterfeiter . LogCacheClient

// LogCacheClient is a client for getting logs.
type LogCacheClient interface {
	Read(
		ctx context.Context,
		sourceID string,
		start time.Time,
		opts ...logcache.ReadOption,
	) ([]*loggregator_v2.Envelope, error)
}
