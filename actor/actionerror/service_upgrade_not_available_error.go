package actionerror

// ServiceUpgradeNotAvailableError is returned when attempting to upgrade a single service instance,
// but there is no upgrade available on the current service plan, i.e., service instance is already
// up-to-date.
type ServiceUpgradeNotAvailableError struct{}

func (e ServiceUpgradeNotAvailableError) Error() string {
	return "No upgrade is available."
}
