package sharedaction

import (
	"context"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
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
