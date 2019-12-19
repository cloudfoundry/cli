package coreconfig

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/util/configv3"
)

type AuthPromptType string

const (
	AuthPromptTypeText     AuthPromptType = "TEXT"
	AuthPromptTypePassword AuthPromptType = "PASSWORD"
)

type AuthPrompt struct {
	Type        AuthPromptType
	DisplayName string
}

type Data struct {
	AccessToken              string
	APIVersion               string
	AsyncTimeout             uint
	AuthorizationEndpoint    string
	ColorEnabled             string
	ConfigVersion            int
	DopplerEndPoint          string
	Locale                   string
	MinCLIVersion            string
	MinRecommendedCLIVersion string
	OrganizationFields       models.OrganizationFields
	PluginRepos              []models.PluginRepo
	RefreshToken             string
	RoutingAPIEndpoint       string
	SpaceFields              models.SpaceFields
	SSHOAuthClient           string
	SSLDisabled              bool
	Target                   string
	Trace                    string
	UaaEndpoint              string
	UAAGrantType             string
	UAAOAuthClient           string
	UAAOAuthClientSecret     string
}

func NewData() *Data {
	data := new(Data)

	data.UAAOAuthClient = "cf"
	data.UAAOAuthClientSecret = ""

	return data
}

func (d *Data) JSONMarshalV3() ([]byte, error) {
	d.ConfigVersion = configv3.CurrentConfigVersion
	return json.MarshalIndent(d, "", "  ")
}

func (d *Data) JSONUnmarshalV3(input []byte) error {
	err := json.Unmarshal(input, d)
	if err != nil {
		return err
	}

	if d.ConfigVersion != configv3.CurrentConfigVersion {
		*d = Data{}
		return nil
	}

	return nil
}
