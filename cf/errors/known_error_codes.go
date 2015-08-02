package errors

const (
	PARSE_ERROR                 = "1001"
	INVALID_RELATION            = "1002"
	NOT_AUTHORIZED              = "10003"
	BAD_QUERY_PARAM             = "10005"
	USER_EXISTS                 = "20002"
	USER_NOT_FOUND              = "20003"
	ORG_EXISTS                  = "30002"
	SPACE_EXISTS                = "40002"
	QUOTA_EXISTS                = "240002"
	SERVICE_INSTANCE_NAME_TAKEN = "60002"
	SERVICE_KEY_NAME_TAKEN      = "360001"
	APP_NOT_STAGED              = "170002"
	APP_STOPPED                 = "220001"
	BUILDPACK_EXISTS            = "290001"
	SECURITY_GROUP_EXISTS       = "300005"
	APP_ALREADY_BOUND           = "90003"
	UNBINDABLE_SERVICE          = "90005"
)
