package actionerror

type MissingSecurityGroupArgsError struct{}

func (MissingSecurityGroupArgsError) Error() string {
	return "Incorrect Usage: the required arguments `SECURITY_GROUP`, `ORG`, and `SPACE` were not provided"
}
