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

type ApplicationsApiResponse struct {
	Resources []ApplicationResource
}

type ApplicationResource struct {
	Metadata Metadata
	Entity   ApplicationEntity
}

type ApplicationEntity struct {
	Name      string
	State     string
	Instances int
	Memory    int
	Routes    []RouteResource
}

type RouteResource struct {
	Metadata Metadata
	Entity   RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain Resource
}
