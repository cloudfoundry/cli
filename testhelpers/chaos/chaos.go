package chaos

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ChaosMonkey provides error injection and chaos testing capabilities
type ChaosMonkey struct {
	failureRate    float64
	latencyMin     time.Duration
	latencyMax     time.Duration
	enabled        bool
	panicRate      float64
	mu             sync.RWMutex
	errorGenerator func() error
}

// NewChaosMonkey creates a new ChaosMonkey instance
func NewChaosMonkey(failureRate float64) *ChaosMonkey {
	rand.Seed(time.Now().UnixNano())

	return &ChaosMonkey{
		failureRate:    failureRate,
		latencyMin:     0,
		latencyMax:     0,
		enabled:        true,
		panicRate:      0,
		errorGenerator: defaultErrorGenerator,
	}
}

// WithLatency configures random latency injection
func (cm *ChaosMonkey) WithLatency(min, max time.Duration) *ChaosMonkey {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.latencyMin = min
	cm.latencyMax = max
	return cm
}

// WithPanicRate configures panic injection rate
func (cm *ChaosMonkey) WithPanicRate(rate float64) *ChaosMonkey {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.panicRate = rate
	return cm
}

// WithCustomErrors sets a custom error generator
func (cm *ChaosMonkey) WithCustomErrors(generator func() error) *ChaosMonkey{
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.errorGenerator = generator
	return cm
}

// Enable enables chaos injection
func (cm *ChaosMonkey) Enable() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.enabled = true
}

// Disable disables chaos injection
func (cm *ChaosMonkey) Disable() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.enabled = false
}

// IsEnabled returns whether chaos injection is enabled
func (cm *ChaosMonkey) IsEnabled() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.enabled
}

// Call wraps a function call with chaos injection
func (cm *ChaosMonkey) Call(fn func() error) error {
	cm.mu.RLock()
	if !cm.enabled {
		cm.mu.RUnlock()
		return fn()
	}

	// Check for panic injection
	if cm.panicRate > 0 && rand.Float64() < cm.panicRate {
		cm.mu.RUnlock()
		panic("Chaos Monkey: Simulated panic!")
	}

	// Inject latency
	if cm.latencyMax > 0 {
		latency := cm.latencyMin + time.Duration(rand.Int63n(int64(cm.latencyMax-cm.latencyMin)))
		cm.mu.RUnlock()
		time.Sleep(latency)
		cm.mu.RLock()
	}

	// Check for error injection
	if rand.Float64() < cm.failureRate {
		err := cm.errorGenerator()
		cm.mu.RUnlock()
		return err
	}

	cm.mu.RUnlock()

	// Call original function
	return fn()
}

// MaybeError returns an error based on failure rate
func (cm *ChaosMonkey) MaybeError() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.enabled {
		return nil
	}

	if rand.Float64() < cm.failureRate {
		return cm.errorGenerator()
	}

	return nil
}

// InjectLatency adds random latency
func (cm *ChaosMonkey) InjectLatency() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.enabled || cm.latencyMax == 0 {
		return
	}

	latency := cm.latencyMin + time.Duration(rand.Int63n(int64(cm.latencyMax-cm.latencyMin)))
	time.Sleep(latency)
}

// MaybePanic potentially panics based on panic rate
func (cm *ChaosMonkey) MaybePanic() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.enabled {
		return
	}

	if rand.Float64() < cm.panicRate {
		panic("Chaos Monkey: Simulated panic!")
	}
}

// Scenario provides a chaos testing scenario
type Scenario struct {
	Name        string
	FailureRate float64
	Latency     time.Duration
	PanicRate   float64
}

// PredefinedScenarios contains common chaos testing scenarios
var PredefinedScenarios = map[string]Scenario{
	"normal": {
		Name:        "Normal Operation",
		FailureRate: 0.0,
		Latency:     0,
		PanicRate:   0.0,
	},
	"network_issues": {
		Name:        "Network Issues",
		FailureRate: 0.3,
		Latency:     100 * time.Millisecond,
		PanicRate:   0.0,
	},
	"high_latency": {
		Name:        "High Latency",
		FailureRate: 0.1,
		Latency:     500 * time.Millisecond,
		PanicRate:   0.0,
	},
	"unstable": {
		Name:        "Unstable Service",
		FailureRate: 0.5,
		Latency:     200 * time.Millisecond,
		PanicRate:   0.1,
	},
	"catastrophic": {
		Name:        "Catastrophic Failure",
		FailureRate: 0.9,
		Latency:     1 * time.Second,
		PanicRate:   0.3,
	},
}

// ApplyScenario applies a predefined chaos scenario
func (cm *ChaosMonkey) ApplyScenario(scenarioName string) error {
	scenario, ok := PredefinedScenarios[scenarioName]
	if !ok {
		return fmt.Errorf("unknown scenario: %s", scenarioName)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.failureRate = scenario.FailureRate
	cm.latencyMin = 0
	cm.latencyMax = scenario.Latency
	cm.panicRate = scenario.PanicRate

	return nil
}

// Default error generator
func defaultErrorGenerator() error {
	errors := []error{
		errors.New("chaos: connection timeout"),
		errors.New("chaos: connection refused"),
		errors.New("chaos: service unavailable"),
		errors.New("chaos: internal server error"),
		errors.New("chaos: rate limit exceeded"),
		errors.New("chaos: authentication failed"),
		errors.New("chaos: resource not found"),
		errors.New("chaos: bad gateway"),
		errors.New("chaos: temporary failure"),
		errors.New("chaos: network unreachable"),
	}

	return errors[rand.Intn(len(errors))]
}

// NetworkChaos simulates network failures
type NetworkChaos struct {
	*ChaosMonkey
}

// NewNetworkChaos creates a chaos monkey for network operations
func NewNetworkChaos() *NetworkChaos {
	return &NetworkChaos{
		ChaosMonkey: NewChaosMonkey(0.2).
			WithLatency(50*time.Millisecond, 500*time.Millisecond).
			WithCustomErrors(func() error {
				networkErrors := []error{
					errors.New("network: connection timeout"),
					errors.New("network: connection refused"),
					errors.New("network: host unreachable"),
					errors.New("network: connection reset"),
					errors.New("network: DNS resolution failed"),
				}
				return networkErrors[rand.Intn(len(networkErrors))]
			}),
	}
}

// APIChaos simulates API failures
type APIChaos struct {
	*ChaosMonkey
}

// NewAPIChaos creates a chaos monkey for API operations
func NewAPIChaos() *APIChaos {
	return &APIChaos{
		ChaosMonkey: NewChaosMonkey(0.15).
			WithLatency(10*time.Millisecond, 200*time.Millisecond).
			WithCustomErrors(func() error {
				apiErrors := []error{
					errors.New("API: 500 Internal Server Error"),
					errors.New("API: 503 Service Unavailable"),
					errors.New("API: 429 Too Many Requests"),
					errors.New("API: 502 Bad Gateway"),
					errors.New("API: 504 Gateway Timeout"),
				}
				return apiErrors[rand.Intn(len(apiErrors))]
			}),
	}
}

// DatabaseChaos simulates database failures
type DatabaseChaos struct {
	*ChaosMonkey
}

// NewDatabaseChaos creates a chaos monkey for database operations
func NewDatabaseChaos() *DatabaseChaos {
	return &DatabaseChaos{
		ChaosMonkey: NewChaosMonkey(0.1).
			WithLatency(5*time.Millisecond, 100*time.Millisecond).
			WithCustomErrors(func() error {
				dbErrors := []error{
					errors.New("database: connection pool exhausted"),
					errors.New("database: deadlock detected"),
					errors.New("database: constraint violation"),
					errors.New("database: connection lost"),
					errors.New("database: query timeout"),
				}
				return dbErrors[rand.Intn(len(dbErrors))]
			}),
	}
}

// Stats tracks chaos monkey statistics
type Stats struct {
	TotalCalls     int
	FailedCalls    int
	PanicCalls     int
	AverageLatency time.Duration
	mu             sync.RWMutex
}

// RecordCall records a chaos monkey call
func (s *Stats) RecordCall(failed bool, panic bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalCalls++
	if failed {
		s.FailedCalls++
	}
	if panic {
		s.PanicCalls++
	}

	// Update average latency
	s.AverageLatency = (s.AverageLatency*time.Duration(s.TotalCalls-1) + latency) / time.Duration(s.TotalCalls)
}

// GetStats returns current statistics
func (s *Stats) GetStats() (int, int, int, time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.TotalCalls, s.FailedCalls, s.PanicCalls, s.AverageLatency
}
