package api

type PaginatedResources struct {
	Resources []Resource
}

type Resource struct {
	Metadata Metadata
	Entity   Entity
}

type Metadata struct {
	Guid string
	Url  string
}

type Entity struct {
	Name string
}
