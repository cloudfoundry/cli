package manifest

type rawManifest struct {
	Applications []Application `yaml:"applications"`

	GlobalBuildpack               interface{} `yaml:"buildpack"`
	GlobalCommand                 interface{} `yaml:"command"`
	GlobalDiskQuota               interface{} `yaml:"disk_quota"`
	GlobalDocker                  interface{} `yaml:"docker"`
	GlobalDomain                  interface{} `yaml:"domain"`
	GlobalDomains                 interface{} `yaml:"domains"`
	GlobalEnv                     interface{} `yaml:"env"`
	GlobalHealthCheckHTTPEndpoint interface{} `yaml:"health-check-http-endpoint"`
	GlobalHealthCheckTimeout      interface{} `yaml:"timeout"`
	GlobalHealthCheckType         interface{} `yaml:"health-check-type"`
	GlobalHost                    interface{} `yaml:"host"`
	GlobalHosts                   interface{} `yaml:"hosts"`
	GlobalInstances               interface{} `yaml:"instances"`
	GlobalMemory                  interface{} `yaml:"memory"`
	GlobalName                    interface{} `yaml:"name"`
	GlobalNoHostname              interface{} `yaml:"no-hostname"`
	GlobalNoRoute                 interface{} `yaml:"no-route"`
	GlobalPath                    interface{} `yaml:"path"`
	GlobalRandomRoute             interface{} `yaml:"random-route"`
	GlobalRoutes                  interface{} `yaml:"routes"`
	GlobalServices                interface{} `yaml:"services"`
	GlobalStack                   interface{} `yaml:"stack"`
	Inherit                       interface{} `yaml:"inherit"`
}

func (raw rawManifest) containsInheritanceField() bool {
	return raw.Inherit != nil
}

func (raw rawManifest) containsGlobalFields() []string {
	globalFields := []string{}

	if raw.GlobalBuildpack != nil {
		globalFields = append(globalFields, "buildpack")
	}
	if raw.GlobalCommand != nil {
		globalFields = append(globalFields, "command")
	}
	if raw.GlobalDiskQuota != nil {
		globalFields = append(globalFields, "disk_quota")
	}
	if raw.GlobalDocker != nil {
		globalFields = append(globalFields, "docker")
	}
	if raw.GlobalDomain != nil {
		globalFields = append(globalFields, "domain")
	}
	if raw.GlobalDomains != nil {
		globalFields = append(globalFields, "domains")
	}
	if raw.GlobalEnv != nil {
		globalFields = append(globalFields, "env")
	}
	if raw.GlobalHealthCheckHTTPEndpoint != nil {
		globalFields = append(globalFields, "health-check-http-endpoint")
	}
	if raw.GlobalHealthCheckTimeout != nil {
		globalFields = append(globalFields, "timeout")
	}
	if raw.GlobalHealthCheckType != nil {
		globalFields = append(globalFields, "health-check-type")
	}
	if raw.GlobalHost != nil {
		globalFields = append(globalFields, "host")
	}
	if raw.GlobalHosts != nil {
		globalFields = append(globalFields, "hosts")
	}
	if raw.GlobalInstances != nil {
		globalFields = append(globalFields, "instances")
	}
	if raw.GlobalMemory != nil {
		globalFields = append(globalFields, "memory")
	}
	if raw.GlobalName != nil {
		globalFields = append(globalFields, "name")
	}
	if raw.GlobalNoHostname != nil {
		globalFields = append(globalFields, "no-hostname")
	}
	if raw.GlobalNoRoute != nil {
		globalFields = append(globalFields, "no-route")
	}
	if raw.GlobalPath != nil {
		globalFields = append(globalFields, "path")
	}
	if raw.GlobalRandomRoute != nil {
		globalFields = append(globalFields, "random-route")
	}
	if raw.GlobalRoutes != nil {
		globalFields = append(globalFields, "routes")
	}
	if raw.GlobalServices != nil {
		globalFields = append(globalFields, "services")
	}
	if raw.GlobalStack != nil {
		globalFields = append(globalFields, "stack")
	}
	return globalFields
}
