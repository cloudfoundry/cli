package load_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cloudfoundry/cli/testhelpers/load"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoad(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Load Testing Suite")
}

var _ = Describe("Load Testing Examples", func() {
	Describe("LoadTester", func() {
		It("performs basic load test", func() {
			// Simulated operation
			operation := func() error {
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				return nil
			}

			// Create load tester: 5 seconds, 10 concurrent users
			tester := load.NewLoadTester(5*time.Second, 10)

			// Run load test
			stats := tester.Run(operation)

			// Verify results
			Expect(stats.RequestCount).To(BeNumerically(">", 0))
			Expect(stats.SuccessRate).To(Equal(100.0))
			Expect(stats.RequestsPerSec).To(BeNumerically(">", 0))

			// Print results
			fmt.Printf("\n📊 Load Test Results:\n")
			fmt.Printf("  Duration:       %v\n", stats.Duration)
			fmt.Printf("  Requests:       %d\n", stats.RequestCount)
			fmt.Printf("  Success Rate:   %.2f%%\n", stats.SuccessRate)
			fmt.Printf("  Requests/sec:   %.2f\n", stats.RequestsPerSec)
			fmt.Printf("  Avg Latency:    %v\n", stats.AvgLatency)
			fmt.Printf("  Min Latency:    %v\n", stats.MinLatency)
			fmt.Printf("  Max Latency:    %v\n", stats.MaxLatency)
			fmt.Printf("  P50:            %v\n", stats.Percentile(50))
			fmt.Printf("  P95:            %v\n", stats.Percentile(95))
			fmt.Printf("  P99:            %v\n", stats.Percentile(99))
		})

		It("performs load test with ramp-up", func() {
			operation := func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}

			// 3 second test, 20 concurrent, 1 second ramp-up
			tester := load.NewLoadTester(3*time.Second, 20).
				WithRampUp(1 * time.Second)

			stats := tester.Run(operation)

			Expect(stats.SuccessRate).To(Equal(100.0))
		})

		It("handles errors correctly", func() {
			errorRate := 0.1 // 10% error rate

			operation := func() error {
				if rand.Float64() < errorRate {
					return fmt.Errorf("simulated error")
				}
				return nil
			}

			tester := load.NewLoadTester(2*time.Second, 5)
			stats := tester.Run(operation)

			// Should have some errors
			Expect(stats.ErrorCount).To(BeNumerically(">", 0))
			Expect(stats.SuccessRate).To(BeNumerically("<", 100.0))
			Expect(stats.SuccessRate).To(BeNumerically(">", 80.0)) // Roughly 90%
		})
	})

	Describe("StressTest", func() {
		It("finds breaking point", func() {
			operation := func() error {
				// Simulated work that degrades with concurrency
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

				// Fail more often under high load
				if rand.Intn(100) > 95 {
					return fmt.Errorf("overloaded")
				}
				return nil
			}

			// Start at 5, max 50, step by 5, 2 seconds per step
			stressTest := load.NewStressTest(5, 50, 5, 2*time.Second)
			results := stressTest.Run(operation)

			fmt.Printf("\n🔥 Stress Test Results:\n")
			for i, stats := range results {
				fmt.Printf("  Step %d (concurrency=%d):\n", i+1, stats.Concurrency)
				fmt.Printf("    Requests/sec:  %.2f\n", stats.RequestsPerSec)
				fmt.Printf("    Success Rate:  %.2f%%\n", stats.SuccessRate)
				fmt.Printf("    Avg Latency:   %v\n", stats.AvgLatency)

				// Mark breaking point
				if stats.SuccessRate < 90.0 {
					fmt.Printf("    ⚠️  BREAKING POINT REACHED\n")
					break
				}
			}

			Expect(len(results)).To(BeNumerically(">", 0))
		})
	})

	Describe("SpikeTest", func() {
		It("tests spike recovery", func() {
			operation := func() error {
				time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				return nil
			}

			// Baseline: 5 concurrent, Spike: 50 concurrent
			spikeTest := load.NewSpikeTest(
				5,              // baseline concurrency
				50,             // spike concurrency
				2*time.Second,  // baseline duration
				1*time.Second,  // spike duration
			)

			baseline, spike, recovery := spikeTest.Run(operation)

			fmt.Printf("\n⚡ Spike Test Results:\n")
			fmt.Printf("  Baseline:\n")
			fmt.Printf("    Requests/sec:  %.2f\n", baseline.RequestsPerSec)
			fmt.Printf("    Avg Latency:   %v\n", baseline.AvgLatency)
			fmt.Printf("\n")
			fmt.Printf("  Spike:\n")
			fmt.Printf("    Requests/sec:  %.2f\n", spike.RequestsPerSec)
			fmt.Printf("    Avg Latency:   %v\n", spike.AvgLatency)
			fmt.Printf("\n")
			fmt.Printf("  Recovery:\n")
			fmt.Printf("    Requests/sec:  %.2f\n", recovery.RequestsPerSec)
			fmt.Printf("    Avg Latency:   %v\n", recovery.AvgLatency)

			// System should recover
			recoveryRatio := recovery.RequestsPerSec / baseline.RequestsPerSec
			Expect(recoveryRatio).To(BeNumerically(">", 0.8)) // At least 80% recovery
		})
	})
})

// Example: Testing CF CLI performance
var _ = Describe("Real-World Load Testing", func() {
	It("tests concurrent app status checks", func() {
		Skip("Example - requires real CF connection")

		operation := func() error {
			// Simulate cf app my-app command
			time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
			return nil
		}

		tester := load.NewLoadTester(10*time.Second, 20)
		stats := tester.Run(operation)

		// Verify performance requirements
		Expect(stats.AvgLatency).To(BeNumerically("<", 200*time.Millisecond))
		Expect(stats.Percentile(95)).To(BeNumerically("<", 500*time.Millisecond))
		Expect(stats.SuccessRate).To(BeNumerically(">", 99.0))
	})

	It("tests route creation under load", func() {
		Skip("Example - requires real CF connection")

		operation := func() error {
			// Simulate cf create-route command
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
			return nil
		}

		// Find breaking point
		stressTest := load.NewStressTest(1, 20, 2, 5*time.Second)
		results := stressTest.Run(operation)

		// Find maximum sustainable throughput
		var maxThroughput float64
		for _, stats := range results {
			if stats.SuccessRate >= 99.0 {
				maxThroughput = stats.RequestsPerSec
			}
		}

		fmt.Printf("Maximum sustainable throughput: %.2f req/sec\n", maxThroughput)
	})
})
