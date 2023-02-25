package helpers

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// GetIsolationSegmentGUID gets the Isolation Segment GUID by passing along the given isolation
// segment name as a query parameter in the /v3/isolation_segments?names=name endpoint.
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
