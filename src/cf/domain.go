package cf

import "fmt"

type InstanceState string

const (
	InstanceStarting InstanceState = "starting"
	InstanceRunning                = "running"
	InstanceFlapping               = "flapping"
	InstanceDown                   = "down"
)

type Organization struct {
	Name string
	Guid string
}

type Space struct {
	Name         string
	Guid         string
	Applications []Application
}

type Application struct {
	Name             string
	Guid             string
	State            string
	Instances        int
	RunningInstances int
	Memory           int
	Urls             []string
	BuildpackUrl     string
	Stack            Stack
}

func (app Application) Health() string {
	if app.State != "started" {
		return app.State
	}

	if app.Instances > 0 {
		ratio := float32(app.RunningInstances) / float32(app.Instances)
		if ratio == 1 {
			return "running"
		}
		return fmt.Sprintf("%.0f%%", ratio*100)
	}

	return "N/A"
}

type Domain struct {
	Name string
	Guid string
}

type Route struct {
	Host string
	Guid string
}

type Stack struct {
	Name string
	Guid string
}

type ApplicationInstance struct {
	State InstanceState
}

type ServicePlan struct {
	Name string
	Guid string
}

type ServiceOffering struct {
	Label string
	Guid  string
	Plans []ServicePlan
}
