package resources

import (
	"code.cloudfoundry.org/cli/types"
)

type ServiceCredentialBindingDetails struct {
	Credentials types.JSONObject `json:"credentials"`
}
