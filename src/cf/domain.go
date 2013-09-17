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
	Name             string
	Guid             string
	Applications     []Application
	ServiceInstances []ServiceInstance
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
	EnvironmentVars  map[string]string
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
	Host   string
	Guid   string
	Domain Domain
}

func (r Route) URL() string {
	return fmt.Sprintf("%s.%s", r.Host, r.Domain.Name)
}

type Stack struct {
	Name        string
	Guid        string
	Description string
}

type ApplicationInstance struct {
	State InstanceState
}

type ServicePlan struct {
	Name            string
	Guid            string
	ServiceOffering ServiceOffering
}

type ServiceOffering struct {
	Guid        string
	Label       string
	Provider    string
	Version     string
	Description string
	Plans       []ServicePlan
}

type ServiceInstance struct {
	Name             string
	Guid             string
	ServiceBindings  []ServiceBinding
	ServicePlan      ServicePlan
	ApplicationNames []string
}

type ServiceBinding struct {
	Url     string
	Guid    string
	AppGuid string
}
