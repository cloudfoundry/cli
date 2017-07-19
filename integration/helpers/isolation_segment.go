package helpers

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GetIsolationSegmentGUID(name string) string {
	session := CF("curl", fmt.Sprintf("/v3/isolation_segments?names=%s", name))
	bytes := session.Wait("15s").Out.Contents()
	return getGUID(bytes)
}

func getGUID(response []byte) string {
	type resource struct {
		Guid string `json:"guid"`
	}
	var GetResponse struct {
		Resources []resource `json:"resources"`
	}

	err := json.Unmarshal(response, &GetResponse)
	Expect(err).ToNot(HaveOccurred())

	if len(GetResponse.Resources) == 0 {
		Fail("No guid found for response")
	}

	return GetResponse.Resources[0].Guid
}
