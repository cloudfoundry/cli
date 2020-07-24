package pushaction

func (actor Actor) BindServices(config ApplicationConfig) (ApplicationConfig, bool, Warnings, error) {
	var allWarnings Warnings
	var boundService bool
	appGUID := config.DesiredApplication.GUID
	for serviceInstanceName, serviceInstance := range config.DesiredServices {
		if _, ok := config.CurrentServices[serviceInstanceName]; !ok {
			warnings, err := actor.V2Actor.BindServiceByApplicationAndServiceInstance(appGUID, serviceInstance.GUID)
			allWarnings = append(allWarnings, warnings...)
			if err != nil {
				return config, false, allWarnings, err
			}
			boundService = true
		}
	}

	config.CurrentServices = config.DesiredServices
	return config, boundService, allWarnings, nil
}
