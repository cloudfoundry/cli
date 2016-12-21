package command_test

import (
	. "code.cloudfoundry.org/cli/command"

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
					"testtokentest": "banana pants"
				}
			}
		}`)

		expected := map[string]interface{}{
			"mytoken": RedactedValue,
			"next_level": map[string]interface{}{
				"next_pAssword_all": RedactedValue,
				"again": map[string]interface{}{
					"real password ": RedactedValue,
					"testtokentest":  RedactedValue,
				},
			},
		}

		redacted, err := SanitizeJSON(raw)
		Expect(err).ToNot(HaveOccurred())
		Expect(redacted).To(Equal(expected))
	})
})
