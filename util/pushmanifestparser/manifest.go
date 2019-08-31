package pushmanifestparser

type Manifest struct {
	Applications   []Application `yaml:"applications"`
	PathToManifest string        `yaml:"-"`
}

func (m Manifest) AppNames() []string {
	var names []string
	for _, app := range m.Applications {
		names = append(names, app.Name)
	}
	return names
}

func (m Manifest) ContainsMultipleApps() bool {
	return len(m.Applications) > 1
}

func (m Manifest) ContainsPrivateDockerImages() bool {
	for _, app := range m.Applications {
		if app.Docker != nil && app.Docker.Username != "" {
			return true
		}
	}
	return false
}

func (m Manifest) GetFirstApp() *Application {
	return &m.Applications[0]
}

func (m Manifest) GetFirstAppWebProcess() *Process {
	for i, process := range m.Applications[0].Processes {
		if process.Type == "web" {
			return &m.Applications[0].Processes[i]
		}
	}

	return nil
}

func (m Manifest) HasAppWithNoName() bool {
	for _, app := range m.Applications {
		if app.Name == "" {
			return true
		}
	}
	return false
}
