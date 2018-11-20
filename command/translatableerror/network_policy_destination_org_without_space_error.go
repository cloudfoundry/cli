package translatableerror

type NetworkPolicyDestinationOrgWithoutSpaceError struct{}

func (NetworkPolicyDestinationOrgWithoutSpaceError) DisplayUsage() {}

func (NetworkPolicyDestinationOrgWithoutSpaceError) Error() string {
	return "Incorrect Usage: space must be provided with org, the required flag '-s' was not specified"
}

func (e NetworkPolicyDestinationOrgWithoutSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
