package helpers

import (
	"os"

	"github.com/cloudfoundry/custom-cats-reporters/honeycomb"
	"github.com/cloudfoundry/custom-cats-reporters/honeycomb/client"
	libhoney "github.com/honeycombio/libhoney-go"
	ginkgo "github.com/onsi/ginkgo"
)

func GetHoneyCombReporter(suiteName string) ginkgo.Reporter {
	writeKey := os.Getenv("HONEYCOMB_WRITE_KEY")
	dataset := os.Getenv("HONEYCOMB_DATASET")
	if writeKey == "" || dataset == "" {
		println("No HONEYCOMB writekey found. Will not log any requests")
		return nil
	}

	honeyCombClient := client.New(libhoney.Config{
		WriteKey: writeKey,
		Dataset:  dataset,
	})

	honeyCombReporter := honeycomb.New(honeyCombClient)
	globalTags := map[string]interface{}{
		"RunID":  os.Getenv("RUN_ID"),
		"API":    GetAPI(),
		"SuitID": suiteName,
	}

	honeyCombReporter.SetGlobalTags(globalTags)
	return honeyCombReporter
}
