package password

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter . PasswordRepository

type PasswordRepository interface {
	UpdatePassword(old string, new string) error
}

type CloudControllerPasswordRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerPasswordRepository(config coreconfig.Reader, gateway net.Gateway) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) error {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return errors.New(T("UAA endpoint missing from config file"))
	}

	url := fmt.Sprintf("/Users/%s/password", repo.config.UserGUID())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(uaaEndpoint, url, strings.NewReader(body))
}
