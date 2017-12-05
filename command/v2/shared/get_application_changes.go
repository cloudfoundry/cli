package shared

import (
	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/util/ui"
)

func GetApplicationChanges(appConfig pushaction.ApplicationConfig) []ui.Change {
	changes := []ui.Change{
		{
			Header:       "name:",
			CurrentValue: appConfig.CurrentApplication.Name,
			NewValue:     appConfig.DesiredApplication.Name,
		},
	}

	if appConfig.DesiredApplication.DockerImage != "" {
		changes = append(changes,
			ui.Change{
				Header:       "docker image:",
				CurrentValue: appConfig.CurrentApplication.DockerImage,
				NewValue:     appConfig.DesiredApplication.DockerImage,
			})

		if appConfig.CurrentApplication.DockerCredentials.Username != "" || appConfig.DesiredApplication.DockerCredentials.Username != "" {
			changes = append(changes,
				ui.Change{
					Header:       "docker username:",
					CurrentValue: appConfig.CurrentApplication.DockerCredentials.Username,
					NewValue:     appConfig.DesiredApplication.DockerCredentials.Username,
				})
		}
	} else {
		changes = append(changes,
			ui.Change{
				Header:       "path:",
				CurrentValue: appConfig.Path,
				NewValue:     appConfig.Path,
			})
	}

	// Existing buildpack and existing detected buildpack are mutually exclusive
	oldBuildpack := appConfig.CurrentApplication.CalculatedBuildpack()
	newBuildpack := appConfig.DesiredApplication.CalculatedBuildpack()
	if oldBuildpack != "" || newBuildpack != "" {
		changes = append(changes,
			ui.Change{
				Header:       "buildpack:",
				CurrentValue: oldBuildpack,
				NewValue:     newBuildpack,
			})
	}

	// Existing command and existing detected start command are mutually exclusive
	oldCommand := appConfig.CurrentApplication.CalculatedCommand()
	newCommand := appConfig.DesiredApplication.CalculatedCommand()
	if oldCommand != "" || newCommand != "" {
		changes = append(changes,
			ui.Change{
				Header:       "command:",
				CurrentValue: oldCommand,
				NewValue:     newCommand,
			})
	}

	if appConfig.CurrentApplication.DiskQuota.IsSet || appConfig.DesiredApplication.DiskQuota.IsSet {
		var currentDiskQuota string
		if appConfig.CurrentApplication.DiskQuota.IsSet {
			currentDiskQuota = appConfig.CurrentApplication.DiskQuota.String()
		}
		changes = append(changes,
			ui.Change{
				Header:       "disk quota:",
				CurrentValue: currentDiskQuota,
				NewValue:     appConfig.DesiredApplication.DiskQuota.String(),
			})
	}

	if appConfig.CurrentApplication.HealthCheckHTTPEndpoint != "" || appConfig.DesiredApplication.HealthCheckHTTPEndpoint != "" {
		changes = append(changes,
			ui.Change{
				Header:       "health check http endpoint:",
				CurrentValue: appConfig.CurrentApplication.HealthCheckHTTPEndpoint,
				NewValue:     appConfig.DesiredApplication.HealthCheckHTTPEndpoint,
			})
	}

	if appConfig.CurrentApplication.HealthCheckTimeout != 0 || appConfig.DesiredApplication.HealthCheckTimeout != 0 {
		changes = append(changes,
			ui.Change{
				Header:       "health check timeout:",
				CurrentValue: appConfig.CurrentApplication.HealthCheckTimeout,
				NewValue:     appConfig.DesiredApplication.HealthCheckTimeout,
			})
	}

	if appConfig.CurrentApplication.HealthCheckType != "" || appConfig.DesiredApplication.HealthCheckType != "" {
		changes = append(changes,
			ui.Change{
				Header:       "health check type:",
				CurrentValue: string(appConfig.CurrentApplication.HealthCheckType),
				NewValue:     string(appConfig.DesiredApplication.HealthCheckType),
			})
	}

	if appConfig.CurrentApplication.Instances.IsSet || appConfig.DesiredApplication.Instances.IsSet {
		changes = append(changes,
			ui.Change{
				Header:       "instances:",
				CurrentValue: appConfig.CurrentApplication.Instances,
				NewValue:     appConfig.DesiredApplication.Instances,
			})
	}

	if appConfig.CurrentApplication.Memory.IsSet || appConfig.DesiredApplication.Memory.IsSet {
		var currentMemory string
		if appConfig.CurrentApplication.Memory.IsSet {
			currentMemory = appConfig.CurrentApplication.Memory.String()
		}
		changes = append(changes,
			ui.Change{
				Header:       "memory:",
				CurrentValue: currentMemory,
				NewValue:     appConfig.DesiredApplication.Memory.String(),
			})
	}

	if appConfig.CurrentApplication.Stack.Name != "" || appConfig.DesiredApplication.Stack.Name != "" {
		changes = append(changes,
			ui.Change{
				Header:       "stack:",
				CurrentValue: appConfig.CurrentApplication.Stack.Name,
				NewValue:     appConfig.DesiredApplication.Stack.Name,
			})
	}

	var oldServices []string
	for name := range appConfig.CurrentServices {
		oldServices = append(oldServices, name)
	}

	var newServices []string
	for name := range appConfig.DesiredServices {
		newServices = append(newServices, name)
	}

	changes = append(changes,
		ui.Change{
			Header:       "services:",
			CurrentValue: oldServices,
			NewValue:     newServices,
		})

	changes = append(changes,
		ui.Change{
			Header:       "env:",
			CurrentValue: appConfig.CurrentApplication.EnvironmentVariables,
			NewValue:     appConfig.DesiredApplication.EnvironmentVariables,
		})

	var currentRoutes []string
	for _, route := range appConfig.CurrentRoutes {
		currentRoutes = append(currentRoutes, route.String())
	}

	var desiredRotues []string
	for _, route := range appConfig.DesiredRoutes {
		desiredRotues = append(desiredRotues, route.String())
	}

	changes = append(changes,
		ui.Change{
			Header:       "routes:",
			CurrentValue: currentRoutes,
			NewValue:     desiredRotues,
		})

	return changes
}

func SelectNonBlankValue(str ...string) string {
	for _, s := range str {
		if s != "" {
			return s
		}
	}
	return ""
}

func MegabytesToString(value uint64) string {
	return bytefmt.ByteSize(bytefmt.MEGABYTE * uint64(value))
}
