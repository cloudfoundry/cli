package performance_test

import (
	"fmt"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("services command performance", func() {
	const (
		serviceName = "service"
		servicePlan = "service-plan"
	)

	var (
		broker           helpers.ServiceBroker
		currentExecution int
		maxExecutions    = getEnvOrDefault("MAX_EXECUTIONS", 10)
		numberOfServices = getEnvOrDefault("NUMBER_OF_SERVICE_INSTANCES", 15)
	)

	BeforeEach(func() {
		helpers.LoginCF()
		helpers.TargetOrgAndSpace(perfOrg, perfSpace)

		currentExecution++
		if os.Getenv("SKIP_PERF_SETUP") == "true" || currentExecution > 1 {
			return
		}

		/* Display some useful information */
		fmt.Printf("Number of samples (MAX_EXECUTIONS): %d\n", maxExecutions)
		fmt.Printf("Number of service instances (NUMBER_OF_SERVICE_INSTANCES): %d\n", numberOfServices)

		domain := helpers.DefaultSharedDomain()
		broker = helpers.CreateBroker(domain, serviceName, servicePlan)

		Eventually(helpers.CF("enable-service-access", serviceName)).Should(Exit(0))

		for i := 0; i < numberOfServices; i++ {
			Eventually(helpers.CF("create-service", serviceName, servicePlan, fmt.Sprintf("instance-%d", i))).Should(Exit(0))
		}
	})

	AfterEach(func() {
		if currentExecution == maxExecutions {
			for i := 0; i < numberOfServices; i++ {
				Eventually(helpers.CF("delete-service", fmt.Sprintf("instance-%d", i), "-f")).Should(Exit(0))
			}
			broker.Destroy()
		}
	})

	Measure("services command", func(b Benchmarker) {
		b.Time("cf services", func() {
			fmt.Printf("cf services...")
			session := helpers.CF("services")
			session.Wait()
			fmt.Printf(" DONE.\n")
			Expect(session).Should(Exit(0))
		})
	}, maxExecutions)
})

func getEnvOrDefault(key string, defaultValue int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	value, err := strconv.Atoi(val)
	if err == nil {
		return value
	}
	return defaultValue
}
