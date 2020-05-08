package ccv3_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
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
				"pagination": {
					"total_results": 0,
					"total_pages": 1,
					"first": {
						"href": "https://fake.com/v3/banana?page=1&per_page=50"
					},
					"last": {
						"href": "https://fake.com/v3/banana?page=2&per_page=50"
					},
					"next": {
						"href":"https://fake.com/v3/banana?page=2&per_page=50"
					},
					"previous": null
				},
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

		It("should populate NextURL", func() {
			Expect(page.NextPage()).To(Equal("https://fake.com/v3/banana?page=2&per_page=50"))
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

	Describe("IncludedResources", func() {
		var raw []byte

		BeforeEach(func() {
			raw = []byte(`{
				"pagination": {
					"next": {
						"href":"https://fake.com/v3/banana?page=2&per_page=50"
					}
				},
				"resources": [
					{
						"metadata": {
							"guid": "app-guid-2",
							"updated_at": null
						},
						"entity": {
							"name": "app-name-2"
						}
					}
				],
				"included": {
      				"users": [
						{
							"guid": "user-guid-1",
							"username": "user-name-1",
							"origin": "uaa"
						}
					],
      				"organizations": [
						{
							"guid": "org-guid-1",
							"name": "org-name-1"
						}
					]
				}
			}`)

			err := json.Unmarshal(raw, &page)
			Expect(err).ToNot(HaveOccurred())
		})

		It("can unmarshal the list of included resources into an appropriate struct", func() {
			Expect(page.IncludedResources.Users).To(ConsistOf(
				resources.User{GUID: "user-guid-1", Username: "user-name-1", Origin: "uaa"},
			))
			Expect(page.IncludedResources.Organizations).To(ConsistOf(
				Organization{GUID: "org-guid-1", Name: "org-name-1"},
			))
		})
	})
})
