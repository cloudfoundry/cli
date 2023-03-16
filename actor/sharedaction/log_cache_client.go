package sharedaction

import (
	"context"
	"time"

	logcache "code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . LogCacheClient

// LogCacheClient is a client for getting logs.
type LogCacheClient interface {
	Read(
		ctx context.Context,
		sourceID string,
		start time.Time,
		opts ...logcache.ReadOption,
	) ([]*loggregator_v2.Envelope, error)
}
