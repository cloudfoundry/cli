package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func Curl(obj interface{}, url string, props ...interface{}) {
	session := CF("curl", fmt.Sprintf(url, props...))
	Eventually(session).Should(Exit(0))
	rawJSON := strings.TrimSpace(string(session.Out.Contents()))

	err := json.Unmarshal([]byte(rawJSON), &obj)
	Expect(err).NotTo(HaveOccurred())
}
