package core_config

import (
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/models"
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
	ConfigVersion            int
	Target                   string
	ApiVersion               string
	AuthorizationEndpoint    string
	LoggregatorEndPoint      string
	DopplerEndPoint          string
	UaaEndpoint              string
	AccessToken              string
	RefreshToken             string
	OrganizationFields       models.OrganizationFields
	SpaceFields              models.SpaceFields
	SSLDisabled              bool
	AsyncTimeout             uint
	Trace                    string
	ColorEnabled             string
	Locale                   string
	PluginRepos              []models.PluginRepo
	MinCliVersion            string
	MinRecommendedCliVersion string
}

func NewData() (data *Data) {
	data = new(Data)
	return
}

func (d *Data) JsonMarshalV3() (output []byte, err error) {
	d.ConfigVersion = 3
	return json.MarshalIndent(d, "", "  ")
}

func (d *Data) JsonUnmarshalV3(input []byte) (err error) {
	err = json.Unmarshal(input, d)
	if err != nil {
		return
	}

	if d.ConfigVersion != 3 {
		*d = Data{}
		return
	}

	return
}
