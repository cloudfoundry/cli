package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
	"strings"
)

type PasswordRepository interface {
	UpdatePassword(old string, new string) error
}

type CloudControllerPasswordRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerPasswordRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) error {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return errors.New(T("UAA endpoint missing from config file"))
	}

	url := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(url, strings.NewReader(body))
}
