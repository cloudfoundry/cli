package load

import (
	"sync"
	"sync/atomic"
	"time"
)

// LoadTester provides load testing capabilities
type LoadTester struct {
	duration       time.Duration
	concurrency    int
	rampUp         time.Duration
	requestCount   int64
	successCount   int64
	errorCount     int64
	totalLatency   int64
	minLatency     int64
	maxLatency     int64
	results        []Result
	mu             sync.Mutex
}

// Result represents a single operation result
type Result struct {
	Latency time.Duration
	Success bool
	Error   error
}

// NewLoadTester creates a new load tester
func NewLoadTester(duration time.Duration, concurrency int) *LoadTester {
	return &LoadTester{
		duration:    duration,
		concurrency: concurrency,
		rampUp:      0,
		minLatency:  int64(^uint64(0) >> 1), // Max int64
		maxLatency:  0,
		results:     make([]Result, 0),
	}
}

// WithRampUp sets the ramp-up period
func (lt *LoadTester) WithRampUp(duration time.Duration) *LoadTester {
	lt.rampUp = duration
	return lt
}

// Run executes the load test
func (lt *LoadTester) Run(operation func() error) *Stats {
	startTime := time.Now()
	endTime := startTime.Add(lt.duration)

	// Channels for coordination
	done := make(chan bool)
	results := make(chan Result, lt.concurrency*100)

	// Worker function
	worker := func(workerID int) {
		// Ramp-up delay
		if lt.rampUp > 0 {
			delay := lt.rampUp * time.Duration(workerID) / time.Duration(lt.concurrency)
			time.Sleep(delay)
		}

		for time.Now().Before(endTime) {
			// Execute operation and measure latency
			opStart := time.Now()
			err := operation()
			latency := time.Since(opStart)

			// Record result
			result := Result{
				Latency: latency,
				Success: err == nil,
				Error:   err,
			}

			results <- result
			atomic.AddInt64(&lt.requestCount, 1)

			if err == nil {
				atomic.AddInt64(&lt.successCount, 1)
			} else {
				atomic.AddInt64(&lt.errorCount, 1)
			}

			// Update latencies
			latencyNs := int64(latency)
			atomic.AddInt64(&lt.totalLatency, latencyNs)

			// Update min/max (with lock for safety)
			lt.mu.Lock()
			if latencyNs < lt.minLatency {
				lt.minLatency = latencyNs
			}
			if latencyNs > lt.maxLatency {
				lt.maxLatency = latencyNs
			}
			lt.mu.Unlock()
		}

		done <- true
	}

	// Start workers
	for i := 0; i < lt.concurrency; i++ {
		go worker(i)
	}

	// Collect results
	go func() {
		for result := range results {
			lt.mu.Lock()
			lt.results = append(lt.results, result)
			lt.mu.Unlock()
		}
	}()

	// Wait for all workers
	for i := 0; i < lt.concurrency; i++ {
		<-done
	}
	close(results)

	// Calculate statistics
	totalTime := time.Since(startTime)
	avgLatency := time.Duration(0)
	if lt.requestCount > 0 {
		avgLatency = time.Duration(lt.totalLatency / lt.requestCount)
	}

	return &Stats{
		Duration:         totalTime,
		RequestCount:     lt.requestCount,
		SuccessCount:     lt.successCount,
		ErrorCount:       lt.errorCount,
		RequestsPerSec:   float64(lt.requestCount) / totalTime.Seconds(),
		AvgLatency:       avgLatency,
		MinLatency:       time.Duration(lt.minLatency),
		MaxLatency:       time.Duration(lt.maxLatency),
		Concurrency:      lt.concurrency,
		SuccessRate:      float64(lt.successCount) / float64(lt.requestCount) * 100,
		Results:          lt.results,
	}
}

// Stats contains load test statistics
type Stats struct {
	Duration        time.Duration
	RequestCount    int64
	SuccessCount    int64
	ErrorCount      int64
	RequestsPerSec  float64
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	Concurrency     int
	SuccessRate     float64
	Results         []Result
}

// Percentile calculates latency percentile
func (s *Stats) Percentile(p float64) time.Duration {
	if len(s.Results) == 0 {
		return 0
	}

	// Sort results by latency
	sorted := make([]time.Duration, len(s.Results))
	for i, r := range s.Results {
		sorted[i] = r.Latency
	}

	// Simple bubble sort (good enough for tests)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * p / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// StressTest gradually increases load until failure
type StressTest struct {
	startConcurrency int
	maxConcurrency   int
	step             int
	stepDuration     time.Duration
}

// NewStressTest creates a new stress tester
func NewStressTest(startConcurrency, maxConcurrency, step int, stepDuration time.Duration) *StressTest {
	return &StressTest{
		startConcurrency: startConcurrency,
		maxConcurrency:   maxConcurrency,
		step:             step,
		stepDuration:     stepDuration,
	}
}

// Run executes the stress test
func (st *StressTest) Run(operation func() error) []Stats {
	var allStats []Stats

	for concurrency := st.startConcurrency; concurrency <= st.maxConcurrency; concurrency += st.step {
		tester := NewLoadTester(st.stepDuration, concurrency)
		stats := tester.Run(operation)
		allStats = append(allStats, *stats)

		// Stop if success rate drops below 90%
		if stats.SuccessRate < 90.0 {
			break
		}
	}

	return allStats
}

// SpikeTest tests system behavior under sudden load spikes
type SpikeTest struct {
	baselineConcurrency int
	spikeConcurrency    int
	baselineDuration    time.Duration
	spikeDuration       time.Duration
}

// NewSpikeTest creates a new spike tester
func NewSpikeTest(baselineConcurrency, spikeConcurrency int, baselineDuration, spikeDuration time.Duration) *SpikeTest {
	return &SpikeTest{
		baselineConcurrency: baselineConcurrency,
		spikeConcurrency:    spikeConcurrency,
		baselineDuration:    baselineDuration,
		spikeDuration:       spikeDuration,
	}
}

// Run executes the spike test
func (st *SpikeTest) Run(operation func() error) (baseline Stats, spike Stats, recovery Stats) {
	// Phase 1: Baseline
	baselineTester := NewLoadTester(st.baselineDuration, st.baselineConcurrency)
	baselineStats := baselineTester.Run(operation)

	// Phase 2: Spike
	spikeTester := NewLoadTester(st.spikeDuration, st.spikeConcurrency)
	spikeStats := spikeTester.Run(operation)

	// Phase 3: Recovery
	recoveryTester := NewLoadTester(st.baselineDuration, st.baselineConcurrency)
	recoveryStats := recoveryTester.Run(operation)

	return *baselineStats, *spikeStats, *recoveryStats
}
