package httpclient

import (
	"net/http"
	"time"

	"github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
)

type retryClient struct {
	delegate              Client
	maxAttempts           uint
	retryDelay            time.Duration
	logger                boshlog.Logger
	isResponseAttemptable func(*http.Response, error) (bool, error)
}

func NewRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
	logger boshlog.Logger,
) Client {
	return &retryClient{
		delegate:              delegate,
		maxAttempts:           maxAttempts,
		retryDelay:            retryDelay,
		logger:                logger,
		isResponseAttemptable: nil,
	}
}

func NewNetworkSafeRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
	logger boshlog.Logger,
) Client {
	return &retryClient{
		delegate:    delegate,
		maxAttempts: maxAttempts,
		retryDelay:  retryDelay,
		logger:      logger,
		isResponseAttemptable: func(resp *http.Response, err error) (bool, error) {
			if err != nil || ((resp.Request.Method == "GET" || resp.Request.Method == "HEAD") && (resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusServiceUnavailable)) {
				return true, errors.WrapError(err, "Retry")
			}

			return false, nil
		},
	}
}

func (r *retryClient) Do(req *http.Request) (*http.Response, error) {
	requestRetryable := NewRequestRetryable(req, r.delegate, r.logger, r.isResponseAttemptable)
	retryStrategy := boshretry.NewAttemptRetryStrategy(int(r.maxAttempts), r.retryDelay, requestRetryable, r.logger)
	err := retryStrategy.Try()

	return requestRetryable.Response(), err
}
