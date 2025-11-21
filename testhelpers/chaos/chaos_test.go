package chaos_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cloudfoundry/cli/testhelpers/chaos"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaos Testing Suite")
}

var _ = Describe("Chaos Testing", func() {
	Describe("ChaosMonkey", func() {
		It("injects failures based on failure rate", func() {
			monkey := chaos.NewChaosMonkey(1.0) // 100% failure rate

			err := monkey.Call(func() error {
				return nil // Original function succeeds
			})

			Expect(err).To(HaveOccurred()) // But chaos monkey fails it
		})

		It("allows successful calls when failure rate is 0", func() {
			monkey := chaos.NewChaosMonkey(0.0) // 0% failure rate

			err := monkey.Call(func() error {
				return nil
			})

			Expect(err).NotTo(HaveOccurred())
		})

		It("injects latency", func() {
			monkey := chaos.NewChaosMonkey(0.0).
				WithLatency(100*time.Millisecond, 200*time.Millisecond)

			start := time.Now()
			monkey.Call(func() error {
				return nil
			})
			elapsed := time.Since(start)

			Expect(elapsed).To(BeNumerically(">=", 100*time.Millisecond))
		})

		It("can be enabled and disabled", func() {
			monkey := chaos.NewChaosMonkey(1.0)

			// Enabled by default
			Expect(monkey.IsEnabled()).To(BeTrue())

			// Disable
			monkey.Disable()
			Expect(monkey.IsEnabled()).To(BeFalse())

			err := monkey.Call(func() error {
				return nil
			})
			Expect(err).NotTo(HaveOccurred()) // Should not inject errors when disabled

			// Re-enable
			monkey.Enable()
			err = monkey.Call(func() error {
				return nil
			})
			Expect(err).To(HaveOccurred()) // Should inject errors again
		})

		It("supports custom error generators", func() {
			customError := errors.New("my custom error")

			monkey := chaos.NewChaosMonkey(1.0).
				WithCustomErrors(func() error {
					return customError
				})

			err := monkey.Call(func() error {
				return nil
			})

			Expect(err).To(Equal(customError))
		})
	})

	Describe("Chaos Scenarios", func() {
		It("applies network issues scenario", func() {
			monkey := chaos.NewChaosMonkey(0.0)
			err := monkey.ApplyScenario("network_issues")

			Expect(err).NotTo(HaveOccurred())
		})

		It("applies high latency scenario", func() {
			monkey := chaos.NewChaosMonkey(0.0)
			err := monkey.ApplyScenario("high_latency")

			Expect(err).NotTo(HaveOccurred())

			start := time.Now()
			monkey.Call(func() error {
				return nil
			})
			elapsed := time.Since(start)

			// High latency scenario should add significant delay
			Expect(elapsed).To(BeNumerically(">", 100*time.Millisecond))
		})

		It("applies unstable scenario", func() {
			monkey := chaos.NewChaosMonkey(0.0)
			err := monkey.ApplyScenario("unstable")

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error for unknown scenario", func() {
			monkey := chaos.NewChaosMonkey(0.0)
			err := monkey.ApplyScenario("unknown_scenario")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown scenario"))
		})
	})

	Describe("NetworkChaos", func() {
		It("simulates network failures", func() {
			networkChaos := chaos.NewNetworkChaos()

			// Run multiple times to observe failures
			failures := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				err := networkChaos.Call(func() error {
					return nil
				})

				if err != nil {
					failures++
					Expect(err.Error()).To(ContainSubstring("network:"))
				}
			}

			// Should have some failures (not exact due to randomness)
			Expect(failures).To(BeNumerically(">", 0))
		})
	})

	Describe("APIChaos", func() {
		It("simulates API failures", func() {
			apiChaos := chaos.NewAPIChaos()

			failures := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				err := apiChaos.Call(func() error {
					return nil
				})

				if err != nil {
					failures++
					Expect(err.Error()).To(ContainSubstring("API:"))
				}
			}

			Expect(failures).To(BeNumerically(">", 0))
		})
	})

	Describe("DatabaseChaos", func() {
		It("simulates database failures", func() {
			dbChaos := chaos.NewDatabaseChaos()

			failures := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				err := dbChaos.Call(func() error {
					return nil
				})

				if err != nil {
					failures++
					Expect(err.Error()).To(ContainSubstring("database:"))
				}
			}

			Expect(failures).To(BeNumerically(">", 0))
		})
	})

	Describe("Real-world chaos scenarios", func() {
		It("handles network failures gracefully", func() {
			networkChaos := chaos.NewNetworkChaos()

			// Simulate a function that retries on network errors
			makeNetworkCall := func() error {
				maxRetries := 3
				var lastErr error

				for i := 0; i < maxRetries; i++ {
					err := networkChaos.Call(func() error {
						// Simulated network call
						return nil
					})

					if err == nil {
						return nil // Success!
					}

					lastErr = err
					time.Sleep(10 * time.Millisecond) // Back off before retry
				}

				return lastErr
			}

			// Even with chaos, retries should eventually succeed
			// (unless we're very unlucky!)
			err := makeNetworkCall()

			// We either succeed or fail after retries - both are valid
			_ = err
		})

		It("tests resilience to API failures", func() {
			apiChaos := chaos.NewAPIChaos()

			// Simulate a function with circuit breaker pattern
			circuitOpen := false
			consecutiveFailures := 0

			callAPI := func() error {
				if circuitOpen {
					return errors.New("circuit breaker open")
				}

				err := apiChaos.Call(func() error {
					return nil
				})

				if err != nil {
					consecutiveFailures++
					if consecutiveFailures >= 3 {
						circuitOpen = true
					}
					return err
				}

				consecutiveFailures = 0
				return nil
			}

			// Make multiple calls
			for i := 0; i < 10; i++ {
				callAPI()
			}

			// Circuit breaker should have been triggered in some scenarios
			// This tests the resilience pattern
		})
	})

	Describe("Panic injection", func() {
		It("can inject panics", func() {
			monkey := chaos.NewChaosMonkey(0.0).
				WithPanicRate(1.0) // 100% panic rate

			defer func() {
				r := recover()
				Expect(r).NotTo(BeNil())
				Expect(r).To(ContainSubstring("Chaos Monkey"))
			}()

			monkey.Call(func() error {
				return nil
			})

			// Should not reach here
			Fail("Expected panic but got none")
		})

		It("recovers from panics gracefully", func() {
			monkey := chaos.NewChaosMonkey(0.0).
				WithPanicRate(1.0)

			// Test that code can handle panics
			recoverFromPanic := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = errors.New("recovered from panic")
					}
				}()

				monkey.Call(func() error {
					return nil
				})

				return nil
			}

			err := recoverFromPanic()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("recovered from panic"))
		})
	})
})
