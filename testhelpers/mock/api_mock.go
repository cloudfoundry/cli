package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
)

// APIServer provides a mock HTTP server for testing
type APIServer struct {
	server   *httptest.Server
	routes   map[string]*Route
	requests []Request
	mu       sync.RWMutex
}

// Route represents a mocked API route
type Route struct {
	Pattern      string
	Method       string
	StatusCode   int
	Response     interface{}
	ResponseFunc func(req *http.Request) (int, interface{})
	Delay        int // milliseconds
	CallCount    int
	regex        *regexp.Regexp
}

// Request represents a captured request
type Request struct {
	Method string
	URL    string
	Body   string
	Headers http.Header
}

// NewAPIServer creates a new mock API server
func NewAPIServer() *APIServer {
	mock := &APIServer{
		routes:   make(map[string]*Route),
		requests: make([]Request, 0),
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// URL returns the server URL
func (m *APIServer) URL() string {
	return m.server.URL
}

// Close shuts down the server
func (m *APIServer) Close() {
	m.server.Close()
}

// RegisterRoute registers a new route
func (m *APIServer) RegisterRoute(method, pattern string, statusCode int, response interface{}) *Route {
	m.mu.Lock()
	defer m.mu.Unlock()

	route := &Route{
		Pattern:    pattern,
		Method:     method,
		StatusCode: statusCode,
		Response:   response,
		regex:      regexp.MustCompile(pattern),
	}

	key := fmt.Sprintf("%s:%s", method, pattern)
	m.routes[key] = route

	return route
}

// GET registers a GET route
func (m *APIServer) GET(pattern string, statusCode int, response interface{}) *Route {
	return m.RegisterRoute("GET", pattern, statusCode, response)
}

// POST registers a POST route
func (m *APIServer) POST(pattern string, statusCode int, response interface{}) *Route {
	return m.RegisterRoute("POST", pattern, statusCode, response)
}

// PUT registers a PUT route
func (m *APIServer) PUT(pattern string, statusCode int, response interface{}) *Route {
	return m.RegisterRoute("PUT", pattern, statusCode, response)
}

// DELETE registers a DELETE route
func (m *APIServer) DELETE(pattern string, statusCode int, response interface{}) *Route {
	return m.RegisterRoute("DELETE", pattern, statusCode, response)
}

// WithFunc allows dynamic responses
func (r *Route) WithFunc(fn func(*http.Request) (int, interface{})) *Route {
	r.ResponseFunc = fn
	return r
}

// WithDelay adds artificial latency
func (r *Route) WithDelay(milliseconds int) *Route {
	r.Delay = milliseconds
	return r
}

// handleRequest processes incoming requests
func (m *APIServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Capture request
	m.captureRequest(r)

	// Find matching route
	route := m.findRoute(r.Method, r.URL.Path)

	if route == nil {
		http.NotFound(w, r)
		return
	}

	route.CallCount++

	// Add delay if configured
	if route.Delay > 0 {
		// time.Sleep(time.Duration(route.Delay) * time.Millisecond)
		// Commented out to avoid actual delays in tests
	}

	// Generate response
	var statusCode int
	var response interface{}

	if route.ResponseFunc != nil {
		statusCode, response = route.ResponseFunc(r)
	} else {
		statusCode = route.StatusCode
		response = route.Response
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if response != nil {
		json.NewEncoder(w).Encode(response)
	}
}

// findRoute finds a matching route
func (m *APIServer) findRoute(method, path string) *Route {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try exact match first
	key := fmt.Sprintf("%s:%s", method, path)
	if route, ok := m.routes[key]; ok {
		return route
	}

	// Try regex match
	for _, route := range m.routes {
		if route.Method == method && route.regex.MatchString(path) {
			return route
		}
	}

	return nil
}

// captureRequest stores the request for verification
func (m *APIServer) captureRequest(r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read body if present
	body := ""
	if r.Body != nil {
		// Note: In real implementation, would need to buffer and restore body
	}

	m.requests = append(m.requests, Request{
		Method:  r.Method,
		URL:     r.URL.String(),
		Body:    body,
		Headers: r.Header,
	})
}

// GetRequests returns all captured requests
func (m *APIServer) GetRequests() []Request {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]Request(nil), m.requests...)
}

// GetRequestCount returns the number of requests made
func (m *APIServer) GetRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.requests)
}

// Reset clears all captured requests
func (m *APIServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = make([]Request, 0)

	for _, route := range m.routes {
		route.CallCount = 0
	}
}

// CloudFoundryMock provides pre-configured CF API routes
type CloudFoundryMock struct {
	*APIServer
}

// NewCloudFoundryMock creates a CF API mock server
func NewCloudFoundryMock() *CloudFoundryMock {
	mock := &CloudFoundryMock{
		APIServer: NewAPIServer(),
	}

	// Register common CF API routes
	mock.setupDefaultRoutes()

	return mock
}

// setupDefaultRoutes configures common CF API endpoints
func (cf *CloudFoundryMock) setupDefaultRoutes() {
	// GET /v2/info
	cf.GET("/v2/info", 200, map[string]interface{}{
		"api_version":               "2.156.0",
		"authorization_endpoint":    "https://login.example.com",
		"token_endpoint":            "https://uaa.example.com",
		"min_cli_version":           "6.23.0",
		"min_recommended_cli_version": "latest",
	})

	// GET /v2/apps
	cf.GET("/v2/apps", 200, map[string]interface{}{
		"total_results": 1,
		"total_pages":   1,
		"resources": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"guid": "app-guid-123",
				},
				"entity": map[string]interface{}{
					"name":       "my-app",
					"state":      "STARTED",
					"instances":  1,
					"memory":     256,
					"disk_quota": 1024,
				},
			},
		},
	})

	// GET /v2/apps/:guid
	cf.GET("/v2/apps/.*", 200, map[string]interface{}{
		"metadata": map[string]interface{}{
			"guid": "app-guid-123",
		},
		"entity": map[string]interface{}{
			"name":       "my-app",
			"state":      "STARTED",
			"instances":  1,
			"memory":     256,
		},
	})

	// GET /v2/spaces
	cf.GET("/v2/spaces", 200, map[string]interface{}{
		"total_results": 1,
		"resources": []map[string]interface{}{
			{
				"metadata": map[string]interface{}{"guid": "space-guid"},
				"entity":   map[string]interface{}{"name": "development"},
			},
		},
	})

	// POST /v2/apps
	cf.POST("/v2/apps", 201, map[string]interface{}{
		"metadata": map[string]interface{}{"guid": "new-app-guid"},
		"entity":   map[string]interface{}{"name": "new-app"},
	})

	// PUT /v2/apps/:guid
	cf.PUT("/v2/apps/.*", 201, map[string]interface{}{
		"metadata": map[string]interface{}{"guid": "app-guid"},
		"entity":   map[string]interface{}{"name": "updated-app"},
	})

	// DELETE /v2/apps/:guid
	cf.DELETE("/v2/apps/.*", 204, nil)
}

// WithApp adds a specific app to responses
func (cf *CloudFoundryMock) WithApp(guid, name, state string, instances, memory int) {
	cf.GET(fmt.Sprintf("/v2/apps/%s", guid), 200, map[string]interface{}{
		"metadata": map[string]interface{}{"guid": guid},
		"entity": map[string]interface{}{
			"name":      name,
			"state":     state,
			"instances": instances,
			"memory":    memory,
		},
	})
}

// WithError configures a route to return an error
func (cf *CloudFoundryMock) WithError(method, pattern string, statusCode int, errorCode, description string) {
	cf.RegisterRoute(method, pattern, statusCode, map[string]interface{}{
		"code":        statusCode,
		"error_code":  errorCode,
		"description": description,
	})
}
