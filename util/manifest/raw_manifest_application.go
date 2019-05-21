package manifest

type rawManifestApplication struct {
	Name                    string             `yaml:"name,omitempty"`
	Buildpack               string             `yaml:"buildpack,omitempty"`
	Buildpacks              []string           `yaml:"buildpacks,omitempty"`
	Command                 string             `yaml:"command,omitempty"`
	DeprecatedDomain        interface{}        `yaml:"domain,omitempty"`
	DeprecatedDomains       interface{}        `yaml:"domains,omitempty"`
	DeprecatedHost          interface{}        `yaml:"host,omitempty"`
	DeprecatedHosts         interface{}        `yaml:"hosts,omitempty"`
	DeprecatedNoHostname    interface{}        `yaml:"no-hostname,omitempty"`
	DiskQuota               string             `yaml:"disk_quota,omitempty"`
	Docker                  rawDockerInfo      `yaml:"docker,omitempty"`
	DropletPath             string             `yaml:"droplet-path,omitempty"`
	EnvironmentVariables    map[string]string  `yaml:"env,omitempty"`
	HealthCheckHTTPEndpoint string             `yaml:"health-check-http-endpoint,omitempty"`
	HealthCheckType         string             `yaml:"health-check-type,omitempty"`
	Instances               *int               `yaml:"instances,omitempty"`
	Memory                  string             `yaml:"memory,omitempty"`
	NoRoute                 bool               `yaml:"no-route,omitempty"`
	Path                    string             `yaml:"path,omitempty"`
	RandomRoute             bool               `yaml:"random-route,omitempty"`
	Routes                  []rawManifestRoute `yaml:"routes,omitempty"`
	Services                []string           `yaml:"services,omitempty"`
	Sidecars                []interface{}      `yaml:"sidecars,omitempty"`
	StackName               string             `yaml:"stack,omitempty"`
	Timeout                 uint64             `yaml:"timeout,omitempty"`
}

type rawManifestRoute struct {
	Route string `yaml:"route"`
}

type rawDockerInfo struct {
	Image    string `yaml:"image,omitempty"`
	Username string `yaml:"username,omitempty"`
}
