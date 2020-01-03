package helpers

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type MetadataLabels map[string]string

func CheckExpectedMetadata(url string, list bool, expected MetadataLabels) {
	type commonResource struct {
		Metadata struct {
			Labels MetadataLabels
		}
	}

	session := CF("curl", url)
	Eventually(session).Should(Exit(0))
	resourceJSON := session.Out.Contents()
	var resource commonResource

	if list {
		var resourceList struct {
			Resources []commonResource
		}

		Expect(json.Unmarshal(resourceJSON, &resourceList)).To(Succeed())
		Expect(resourceList.Resources).To(HaveLen(1))
		resource = resourceList.Resources[0]
	} else {
		Expect(json.Unmarshal(resourceJSON, &resource)).To(Succeed())
	}

	Expect(resource.Metadata.Labels).To(HaveLen(len(expected)))
	for k, v := range expected {
		Expect(resource.Metadata.Labels).To(HaveKeyWithValue(k, v))
	}
}
