package v3action

func (actor Actor) GetAppScaleSummaryByNameAndSpace(name string, spaceGUID string) (AppScaleSummary, Warnings, error) {
	return AppScaleSummary{}, Warnings{}, nil
}

func (actor Actor) UpdateAppScale(name string, spaceGUID string, numInstances int, memUsage int, diskUsage int) (AppScaleSummary, Warnings, error) {
	return AppScaleSummary{}, Warnings{}, nil
}
