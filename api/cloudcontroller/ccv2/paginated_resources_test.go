package ccv2_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type testItem struct {
	Name string
	GUID string
}

func (t *testItem) UnmarshalJSON(data []byte) error {
	var item struct {
		Metadata struct {
			GUID string `json:"guid"`
		} `json:"metadata"`
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &item)
	if err != nil {
		return err
	}

	t.GUID = item.Metadata.GUID
	t.Name = item.Entity.Name
	return nil
}

var _ = Describe("Paginated Resources", func() {
	var page *PaginatedResources

	BeforeEach(func() {
		page = NewPaginatedResources(testItem{})
	})

	Context("unmarshaling from paginated request", func() {
		var raw []byte

		BeforeEach(func() {
			raw = []byte(`{
				"next_url": "https://no-idea/some-cc-url&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]
			}`)

			err := json.Unmarshal(raw, &page)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should populate the next_url", func() {
			Expect(page.NextURL).To(Equal("https://no-idea/some-cc-url&page=2"))
		})

		It("should hold onto the whole resource blob", func() {
			Expect(string(page.ResourcesBytes)).To(MatchJSON(`[
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]`))
		})
	})

	Describe("Resources", func() {
		BeforeEach(func() {
			raw := []byte(`[
					{
						"metadata": {
							"guid": "app-guid-1",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-1"
						}
					},
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				]`)

			page.ResourcesBytes = raw
		})

		It("can unmarshal the list of resources into the given struct", func() {
			items, err := page.Resources()
			Expect(err).ToNot(HaveOccurred())

			Expect(items).To(ConsistOf(
				testItem{GUID: "app-guid-1", Name: "app-name-1"},
				testItem{GUID: "app-guid-2", Name: "app-name-2"},
			))
		})
	})
})
