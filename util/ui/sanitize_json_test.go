package ui_test

import (
	"strings"

	. "code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SanitizeJSON", func() {
	DescribeTable("sanitizes the json input",
		func(original string) {
			expected := strings.Replace(original, "CrAzY_PaSSw0rd", RedactedValue, -1)

			redacted, err := SanitizeJSON([]byte(original))
			Expect(err).ToNot(HaveOccurred())
			Expect(redacted).To(MatchJSON(expected))
		},

		Entry("when the top level is an array", `[
			{
				"some url":"jdbc:mysql://hostname/db-name?user=username&password=CrAzY_PaSSw0rd",
				"some other url":"jdbc:mysql://hostname/db-name?password=CrAzY_PaSSw0rd&user=username",
				"uri":"postgres://some-user-name:CrAzY_PaSSw0rd@10.0.0.1:5432/some-other-data"
			},
			{
			"real password ": "CrAzY_PaSSw0rd",
			"testtokentest": "CrAzY_PaSSw0rd",
			"simple": "https://www.google.com/search?q=i+am+a+potato&oq=I+am+a+potato&aqs=chrome.0.0l6.2383j0j8&client=ubuntu&sourceid=chrome&ie=UTF-8"
			}
		]`),

		Entry("when the top level is an array", `{
			"mytoken": "CrAzY_PaSSw0rd",
			"next_level": {
				"again": {
					"real password ": "CrAzY_PaSSw0rd",
					"simple": "https://www.google.com/search?q=i+am+a+potato&oq=I+am+a+potato&aqs=chrome.0.0l6.2383j0j8&client=ubuntu&sourceid=chrome&ie=UTF-8",
					"some other url":"jdbc:mysql://hostname/db-name?password=CrAzY_PaSSw0rd&user=username",
					"some url":"jdbc:mysql://hostname/db-name?user=username&password=CrAzY_PaSSw0rd",
					"testtokentest": "CrAzY_PaSSw0rd",
					"uri":"postgres://some-user-name:CrAzY_PaSSw0rd@10.0.0.1:5432/some-other-data"
				},
				"ary": [
					"jdbc:mysql://hostname/db-name?user=username&password=CrAzY_PaSSw0rd",
					"postgres://some-user-name:CrAzY_PaSSw0rd@10.0.0.1:5432/some-other-data"
				],
				"ary2": [
					{
						"some other url":"jdbc:mysql://hostname/db-name?password=CrAzY_PaSSw0rd&user=username",
						"some url":"jdbc:mysql://hostname/db-name?user=username&password=CrAzY_PaSSw0rd",
						"uri":"postgres://some-user-name:CrAzY_PaSSw0rd@10.0.0.1:5432/some-other-data"
					}
				],
				"next_pAssword_all": "CrAzY_PaSSw0rd"
			}
		}`),
	)

	It("formats spacing", func() {
		original := `{"a":"b"}`
		expected := `{
  "a": "b"
}
` // Extra new line is required due to encoder
		redacted, err := SanitizeJSON([]byte(original))
		Expect(err).ToNot(HaveOccurred())
		Expect(redacted).To(Equal([]byte(expected)))
	})

	It("does not escape characters", func() {
		original := `{"a":"&<foo#>"}`
		expected := `{
  "a": "&<foo#>"
}
` // Extra new line is required due to encoder
		redacted, err := SanitizeJSON([]byte(original))
		Expect(err).ToNot(HaveOccurred())
		Expect(redacted).To(Equal([]byte(expected)))
	})

	It("represents empty arrays as []", func() {
		original := `{"resources": []}`
		expected := `{"resources": []}
` // Extra new line is required due to encoder

		redacted, err := SanitizeJSON([]byte(original))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(redacted)).To(MatchJSON(expected))
	})
})
