package main

import (
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/metrics"
	"time"
)

var appId = "60a13b0f-fce7-4c02-b92a-d43d583877ed"

func main() {
	err := dropsonde.Initialize("localhost:3457", "METRIC-TEST", "z1", "0")
	if err != nil {
		println(err.Error())
	}

	var i uint64
	i = 0
	for {
		println("emitting metric at counter: ", i)
		metrics.SendContainerMetric(appId, 0, 42.42, 1234, i)
		metrics.SendContainerMetric(appId, 1, 11.41, 1234, i)
		metrics.SendContainerMetric(appId, 2, 11.41, 1234, i)
		metrics.SendContainerMetric("donotseethis", 2, 11.41, 1234, i)
		i++
		time.Sleep(1 * time.Second)
	}
}
