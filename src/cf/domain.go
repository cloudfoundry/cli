package cf

type Organization struct {
	Name string
	Guid string
}

type Space struct {
	Name string
	Guid string
}

type Application struct {
	Name      string
	Guid      string
	State     string
	Instances int
	Memory    int
	Urls      []string
}

type Domain struct {
	Name string
	Guid string
}

type Route struct {
	Host string
	Guid string
}

type ApplicationInstance struct {
	State string
}
