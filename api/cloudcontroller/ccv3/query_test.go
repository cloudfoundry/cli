package ccv3_test

import (
	"net/url"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Query Helpers", func() {
	Describe("FormatQueryParameters", func() {
		var (
			inputQueries     []Query
			outputParameters url.Values
		)

		BeforeEach(func() {
			inputQueries = []Query{
				{
					Key:    SpaceGUIDFilter,
					Values: []string{"space-guid1", "space-guid2"},
				},
			}
			outputParameters = FormatQueryParameters(inputQueries)
		})

		It("encodes the param values and reformats the query", func() {
			Expect(outputParameters).To(Equal(url.Values{
				"space_guids": []string{"space-guid1,space-guid2"},
			}))
		})

		When("the name filter is used", func() {
			BeforeEach(func() {
				inputQueries = []Query{
					{
						Key:    NameFilter,
						Values: []string{"name1", "name,2", "name 3"},
					},
				}
				outputParameters = FormatQueryParameters(inputQueries)
			})

			It("encodes commas before formatting", func() {
				Expect(outputParameters).To(Equal(url.Values{
					"names": []string{"name1,name%2C2,name 3"},
				}))
			})
		})
	})
})
