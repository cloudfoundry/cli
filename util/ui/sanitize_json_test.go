package ui_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SanitizeJSON", func() {
	It("sanitizes json", func() {
		raw := []byte(`{
			"mytoken": "foo",
			"next_level": {
				"next_pAssword_all": "bar",
				"again": {
					"real password ": "Don't tell nobody, it's banana",
					"token_endpoint": "some url",
					"testtokentest": "banana pants",
					"some url":"jdbc:mysql://hostname/db-name?user=username&password=very-secret-password",
					"some other url":"jdbc:mysql://hostname/db-name?password=very-secret-password&user=username"
				}
			}
		}`)

		expected := map[string]interface{}{
			"mytoken": RedactedValue,
			"next_level": map[string]interface{}{
				"next_pAssword_all": RedactedValue,
				"again": map[string]interface{}{
					"real password ": RedactedValue,
					"token_endpoint": "some url",
					"testtokentest":  RedactedValue,
					"some url":       fmt.Sprintf("jdbc:mysql://hostname/db-name?user=username&password=%s", RedactedValue),
					"some other url": fmt.Sprintf("jdbc:mysql://hostname/db-name?password=%s&user=username", RedactedValue),
				},
			},
		}

		redacted, err := SanitizeJSON(raw)
		Expect(err).ToNot(HaveOccurred())
		Expect(redacted).To(Equal(expected))
	})
})
