package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

// Curl runs a 'cf curl' command with a URL format string, allowing props to be
// interpolated into the URL string with fmt.Sprintf. The JSON response is
// unmarshaled into given obj.
func Curl(obj interface{}, url string, props ...interface{}) {
	session := CF("curl", fmt.Sprintf(url, props...))
	Eventually(session).Should(Exit(0))
	rawJSON := strings.TrimSpace(string(session.Out.Contents()))

	err := json.Unmarshal([]byte(rawJSON), &obj)
	Expect(err).NotTo(HaveOccurred())
}
