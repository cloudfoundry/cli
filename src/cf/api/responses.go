package api

type Metadata struct {
	Guid string
}

type Entity struct {
	Name string
	Host string
}

type Resource struct {
	Metadata Metadata
	Entity   Entity
}

type ApiResponse struct {
	Resources []Resource
}
