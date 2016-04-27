package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/metrics"
)

var (
	appId      = os.Getenv("APP_GUID")
	metronAddr = defaultEnv("METRON_ADDR", "localhost:3457")
)

func main() {
	err := dropsonde.Initialize(metronAddr, "METRIC-TEST", "z1", "0")
	if err != nil {
		println(err.Error())
	}

	for i := uint64(0); ; i++ {
		println("emitting metric at counter: ", i)
		metrics.SendContainerMetric(appId, 0, 42.42, 1234, i)
		metrics.SendContainerMetric(appId, 1, 11.41, 1234, i)
		metrics.SendContainerMetric(appId, 2, 11.41, 1234, i)
		metrics.SendContainerMetric("donotseethis", 2, 11.41, 1234, i)
		time.Sleep(1 * time.Second)
	}
}

func defaultEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
