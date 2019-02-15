package password

import (
	"bytes"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/net"
)

//go:generate counterfeiter . Repository

type Repository interface {
	UpdatePassword(old string, new string) error
}

type CloudControllerRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerRepository(config coreconfig.Reader, gateway net.Gateway) (repo CloudControllerRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerRepository) UpdatePassword(old string, new string) error {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return errors.New(T("UAA endpoint missing from config file"))
	}

	type passwordReq struct {
		Old string `json:"oldPassword"`
		New string `json:"password"`
	}

	url := fmt.Sprintf("/Users/%s/password", repo.config.UserGUID())
	reqBody := passwordReq{
		Old: old,
		New: new,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	return repo.gateway.UpdateResource(uaaEndpoint, url, bytes.NewReader(bodyBytes))
}
