package translatableerror

type NetworkPolicyDestinationOrgWithoutSpaceError struct{}

func (NetworkPolicyDestinationOrgWithoutSpaceError) DisplayUsage() {}

func (NetworkPolicyDestinationOrgWithoutSpaceError) Error() string {
	return "Incorrect Usage: A space must be provided when an org is provided, but the '-s' flag was not specified."
}

func (e NetworkPolicyDestinationOrgWithoutSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
