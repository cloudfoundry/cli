package manifest

type rawManifest struct {
	Applications []Application `yaml:"applications"`

	DeprecatedBuildpack               interface{} `yaml:"buildpack"`
	DeprecatedCommand                 interface{} `yaml:"command"`
	DeprecatedDiskQuota               interface{} `yaml:"disk_quota"`
	DeprecatedDocker                  interface{} `yaml:"docker"`
	DeprecatedDomain                  interface{} `yaml:"domain"`
	DeprecatedDomains                 interface{} `yaml:"domains"`
	DeprecatedEnv                     interface{} `yaml:"env"`
	DeprecatedHealthCheckHTTPEndpoint interface{} `yaml:"health-check-http-endpoint"`
	DeprecatedHealthCheckTimeout      interface{} `yaml:"timeout"`
	DeprecatedHealthCheckType         interface{} `yaml:"health-check-type"`
	DeprecatedHost                    interface{} `yaml:"host"`
	DeprecatedHosts                   interface{} `yaml:"hosts"`
	DeprecatedInherit                 interface{} `yaml:"inherit"`
	DeprecatedInstances               interface{} `yaml:"instances"`
	DeprecatedMemory                  interface{} `yaml:"memory"`
	DeprecatedName                    interface{} `yaml:"name"`
	DeprecatedNoHostname              interface{} `yaml:"no-hostname"`
	DeprecatedNoRoute                 interface{} `yaml:"no-route"`
	DeprecatedPath                    interface{} `yaml:"path"`
	DeprecatedRandomRoute             interface{} `yaml:"random-route"`
	DeprecatedRoutes                  interface{} `yaml:"routes"`
	DeprecatedServices                interface{} `yaml:"services"`
	DeprecatedStack                   interface{} `yaml:"stack"`
}

func (raw rawManifest) containsDeprecatedFields() bool {
	return raw.DeprecatedBuildpack != nil ||
		raw.DeprecatedCommand != nil ||
		raw.DeprecatedDiskQuota != nil ||
		raw.DeprecatedDocker != nil ||
		raw.DeprecatedDomain != nil ||
		raw.DeprecatedDomains != nil ||
		raw.DeprecatedEnv != nil ||
		raw.DeprecatedHealthCheckHTTPEndpoint != nil ||
		raw.DeprecatedHealthCheckTimeout != nil ||
		raw.DeprecatedHealthCheckType != nil ||
		raw.DeprecatedHost != nil ||
		raw.DeprecatedHosts != nil ||
		raw.DeprecatedInherit != nil ||
		raw.DeprecatedInstances != nil ||
		raw.DeprecatedMemory != nil ||
		raw.DeprecatedName != nil ||
		raw.DeprecatedNoHostname != nil ||
		raw.DeprecatedNoRoute != nil ||
		raw.DeprecatedPath != nil ||
		raw.DeprecatedRandomRoute != nil ||
		raw.DeprecatedRoutes != nil ||
		raw.DeprecatedServices != nil ||
		raw.DeprecatedStack != nil
}
