package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/log-cache/pkg/marshaler"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	"github.com/blang/semver"
	"github.com/golang/protobuf/jsonpb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// Client reads from LogCache via the RESTful or gRPC API.
type Client struct {
	addr        string
	baseApiPath string

	httpClient       HTTPClient
	grpcClient       logcache_v1.EgressClient
	promqlGrpcClient logcache_v1.PromQLQuerierClient
}

// NewIngressClient creates a Client.
func NewClient(addr string, opts ...ClientOption) *Client {
	c := &Client{
		addr: addr,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	for _, o := range opts {
		o.configure(c)
	}

	return c
}

// ClientOption configures the LogCache client.
type ClientOption interface {
	configure(client interface{})
}

// clientOptionFunc enables regular functions to be a ClientOption.
type clientOptionFunc func(client interface{})

// configure Implements clientOptionFunc.
func (f clientOptionFunc) configure(client interface{}) {
	f(client)
}

// HTTPClient is an interface that represents a http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WithHTTPClient sets the HTTP client. It defaults to a client that timesout
// after 5 seconds.
func WithHTTPClient(h HTTPClient) ClientOption {
	return clientOptionFunc(func(c interface{}) {
		switch c := c.(type) {
		case *Client:
			c.httpClient = h
		default:
			panic("unknown type")
		}
	})
}

// WithViaGRPC enables gRPC instead of HTTP/1 for reading from LogCache.
func WithViaGRPC(opts ...grpc.DialOption) ClientOption {
	return clientOptionFunc(func(c interface{}) {
		switch c := c.(type) {
		case *Client:
			conn, err := grpc.Dial(c.addr, opts...)
			if err != nil {
				panic(fmt.Sprintf("failed to dial via gRPC: %s", err))
			}

			c.grpcClient = logcache_v1.NewEgressClient(conn)
			c.promqlGrpcClient = logcache_v1.NewPromQLQuerierClient(conn)
		default:
			panic("unknown type")
		}
	})
}

// Read queries the LogCache and returns the given envelopes. To override any
// query defaults (e.g., end time), use the according option.
func (c *Client) Read(
	ctx context.Context,
	sourceID string,
	start time.Time,
	opts ...ReadOption,
) ([]*loggregator_v2.Envelope, error) {
	if c.grpcClient != nil {
		return c.grpcRead(ctx, sourceID, start, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}

	baseApiPath, err := c.getBaseApiPath(ctx)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("%s/read/%s", baseApiPath, sourceID)
	q := u.Query()
	q.Set("start_time", strconv.FormatInt(start.UnixNano(), 10))

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var r logcache_v1.ReadResponse
	if err := jsonpb.Unmarshal(resp.Body, &r); err != nil {
		return nil, err
	}

	return r.GetEnvelopes().GetBatch(), nil
}

// ReadOption configures the URL that is used to submit the query. The
// RawQuery is set to the decoded query parameters after each option is
// invoked.
type ReadOption func(u *url.URL, q url.Values)

// WithEndTime sets the 'end_time' query parameter to the given time. It
// defaults to empty, and therefore the end of the cache.
func WithEndTime(t time.Time) ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("end_time", strconv.FormatInt(t.UnixNano(), 10))
	}
}

// WithLimit sets the 'limit' query parameter to the given value. It
// defaults to empty, and therefore 100 envelopes.
func WithLimit(limit int) ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("limit", strconv.Itoa(limit))
	}
}

// WithEnvelopeTypes sets the 'envelope_types' query parameter to the given
// value. It defaults to empty, and therefore any envelope type.
func WithEnvelopeTypes(t ...logcache_v1.EnvelopeType) ReadOption {
	return func(u *url.URL, q url.Values) {
		for _, v := range t {
			q.Add("envelope_types", v.String())
		}
	}
}

// WithDescending set the 'descending' query parameter to true. It defaults to
// false, yielding ascending order.
func WithDescending() ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("descending", "true")
	}
}

func WithNameFilter(nameFilter string) ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("name_filter", nameFilter)
	}
}

func (c *Client) grpcRead(ctx context.Context, sourceID string, start time.Time, opts []ReadOption) ([]*loggregator_v2.Envelope, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.ReadRequest{
		SourceId:  sourceID,
		StartTime: start.UnixNano(),
	}

	if v, ok := q["limit"]; ok {
		req.Limit, _ = strconv.ParseInt(v[0], 10, 64)
	}

	if v, ok := q["end_time"]; ok {
		req.EndTime, _ = strconv.ParseInt(v[0], 10, 64)
	}

	for _, et := range q["envelope_types"] {
		req.EnvelopeTypes = append(req.EnvelopeTypes,
			logcache_v1.EnvelopeType(logcache_v1.EnvelopeType_value[et]),
		)
	}

	if v, ok := q["name_filter"]; ok {
		req.NameFilter = v[0]
	}

	if _, ok := q["descending"]; ok {
		req.Descending = true
	}

	resp, err := c.grpcClient.Read(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Envelopes.Batch, nil
}

// Meta returns meta information from the entire LogCache.
func (c *Client) Meta(ctx context.Context) (map[string]*logcache_v1.MetaInfo, error) {
	if c.grpcClient != nil {
		return c.grpcMeta(ctx)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}

	baseApiPath, err := c.getBaseApiPath(ctx)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("%s/meta", baseApiPath)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var metaResponse logcache_v1.MetaResponse
	if err := jsonpb.Unmarshal(resp.Body, &metaResponse); err != nil {
		return nil, err
	}

	return metaResponse.Meta, nil
}

func (c *Client) grpcMeta(ctx context.Context) (map[string]*logcache_v1.MetaInfo, error) {
	resp, err := c.grpcClient.Meta(ctx, &logcache_v1.MetaRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Meta, nil
}

func (c *Client) getBaseApiPath(ctx context.Context) (string, error) {
	if c.baseApiPath != "" {
		return c.baseApiPath, nil
	}

	logCacheVersion, err := c.LogCacheVersion(ctx)
	if err != nil {
		return "", err
	}

	if logCacheVersion.GTE(FIRST_LOG_CACHE_VERSION_AFTER_API_MOVE) {
		return "/api/v1", nil
	}

	return "/v1", nil
}

func (c *Client) LogCacheVersion(ctx context.Context) (semver.Version, error) {
	u, err := url.Parse(c.addr)
	if err != nil {
		return semver.Version{}, err
	}

	u.Path = "/api/v1/info"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return semver.Version{}, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return semver.Version{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return LAST_LOG_CACHE_VERSION_WITHOUT_INFO, nil
	}

	if resp.StatusCode != http.StatusOK {
		return semver.Version{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var info struct {
		Version string `json:"version"`
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(info.Version)
}

func (c *Client) LogCacheVMUptime(ctx context.Context) (int64, error) {
	u, err := url.Parse(c.addr)
	if err != nil {
		return -1, err
	}

	u.Path = "/api/v1/info"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var info struct {
		VMUptime string `json:"vm_uptime"`
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return -1, err
	}

	if info.VMUptime == "" {
		return -1, errors.New("This version of log cache does not support vm_uptime info")
	}

	uptime, err := strconv.ParseInt(info.VMUptime, 10, 64)
	if err != nil {
		return -1, err
	}

	return uptime, nil
}

// PromQLOption configures the URL that is used to submit the query. The
// RawQuery is set to the decoded query parameters after each option is
// invoked.
type PromQLOption func(u *url.URL, q url.Values)

// WithPromQLTime returns a PromQLOption that configures the 'time' query
// parameter for a PromQL query.
func WithPromQLTime(t time.Time) PromQLOption {
	return func(u *url.URL, q url.Values) {
		q.Set("time", formatDecimalTimeWithMillis(t))
	}
}

func WithPromQLStart(t time.Time) PromQLOption {
	return func(u *url.URL, q url.Values) {
		q.Set("start", formatDecimalTimeWithMillis(t))
	}
}

func WithPromQLEnd(t time.Time) PromQLOption {
	return func(u *url.URL, q url.Values) {
		q.Set("end", formatDecimalTimeWithMillis(t))
	}
}

func formatDecimalTimeWithMillis(t time.Time) string {
	return fmt.Sprintf("%.3f", float64(t.UnixNano())/1e9)
}

func WithPromQLStep(step string) PromQLOption {
	return func(u *url.URL, q url.Values) {
		q.Set("step", step)
	}
}

// PromQL issues a PromQL range query against Log Cache data.
func (c *Client) PromQLRange(
	ctx context.Context,
	query string,
	opts ...PromQLOption,
) (*logcache_v1.PromQL_RangeQueryResult, error) {
	if c.promqlGrpcClient != nil {
		return c.grpcPromQLRange(ctx, query, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", query)

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var promQLResponse logcache_v1.PromQL_RangeQueryResult
	marshaler := marshaler.NewPromqlMarshaler(&runtime.JSONPb{})
	if err := marshaler.NewDecoder(resp.Body).Decode(&promQLResponse); err != nil {
		return nil, err
	}

	return &promQLResponse, nil
}

func (c *Client) grpcPromQLRange(ctx context.Context, query string, opts []PromQLOption) (*logcache_v1.PromQL_RangeQueryResult, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.PromQL_RangeQueryRequest{
		Query: query,
	}

	if v, ok := q["start"]; ok {
		req.Start = v[0]
	}

	if v, ok := q["end"]; ok {
		req.End = v[0]
	}

	if v, ok := q["step"]; ok {
		req.Step = v[0]
	}

	resp, err := c.promqlGrpcClient.RangeQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) PromQLRangeRaw(
	ctx context.Context,
	query string,
	opts ...PromQLOption,
) (*PromQLQueryResult, error) {
	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/query_range"
	q := u.Query()
	q.Set("query", query)

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PromQLQueryResult

	// If we got a 404, it's probably due to lack of authorization. Let's try
	// to be nice to users and give them a hint.
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code %d (check authorization?)", resp.StatusCode)
	}

	// The PromQL API will return nicely-formatted JSON errors with a
	// status code of either 400 (Bad Request) or 500 (Internal Server Error),
	// but we can return a generic message for every other status code.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusBadRequest &&
		resp.StatusCode != http.StatusInternalServerError {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%s (status code %d)", err.Error(), resp.StatusCode)
	}

	return &result, nil
}

// PromQL issues a PromQL instant query against Log Cache data.
func (c *Client) PromQL(
	ctx context.Context,
	query string,
	opts ...PromQLOption,
) (*logcache_v1.PromQL_InstantQueryResult, error) {
	if c.promqlGrpcClient != nil {
		return c.grpcPromQL(ctx, query, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", query)

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var promQLResponse logcache_v1.PromQL_InstantQueryResult
	marshaler := marshaler.NewPromqlMarshaler(&runtime.JSONPb{})
	if err := marshaler.NewDecoder(resp.Body).Decode(&promQLResponse); err != nil {
		return nil, err
	}

	return &promQLResponse, nil
}

func (c *Client) grpcPromQL(ctx context.Context, query string, opts []PromQLOption) (*logcache_v1.PromQL_InstantQueryResult, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.PromQL_InstantQueryRequest{
		Query: query,
	}

	if v, ok := q["time"]; ok {
		req.Time = v[0]
	}

	resp, err := c.promqlGrpcClient.InstantQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) PromQLRaw(
	ctx context.Context,
	query string,
	opts ...PromQLOption,
) (*PromQLQueryResult, error) {
	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "/api/v1/query"
	q := u.Query()
	q.Set("query", query)

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PromQLQueryResult

	// If we got a 404, it's probably due to lack of authorization. Let's try
	// to be nice to users and give them a hint.
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("unexpected status code %d (check authorization?)", resp.StatusCode)
	}

	// The PromQL API will return nicely-formatted JSON errors with a
	// status code of either 400 (Bad Request) or 500 (Internal Server Error),
	// but we can return a generic message for every other status code.
	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusBadRequest &&
		resp.StatusCode != http.StatusInternalServerError {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%s (status code %d)", err.Error(), resp.StatusCode)
	}

	return &result, nil
}

type PromQLQueryResult struct {
	Status    string           `json:"status"`
	Data      PromQLResultData `json:"data"`
	ErrorType string           `json:"errorType,omitempty"`
	Error     string           `json:"error,omitempty"`
}

type PromQLResultData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result,omitempty"`
}

var (
	LAST_LOG_CACHE_VERSION_WITHOUT_INFO = semver.Version{
		Major: 1,
		Minor: 4,
		Patch: 7,
	}
	FIRST_LOG_CACHE_VERSION_AFTER_API_MOVE = semver.Version{
		Major: 2,
		Minor: 0,
		Patch: 0,
	}
)
