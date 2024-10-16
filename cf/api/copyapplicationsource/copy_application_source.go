package copyapplicationsource

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/net"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Repository

type Repository interface {
	CopyApplication(sourceAppGUID, targetAppGUID string) error
}

type CloudControllerApplicationSourceRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerCopyApplicationSourceRepository(config coreconfig.Reader, gateway net.Gateway) *CloudControllerApplicationSourceRepository {
	return &CloudControllerApplicationSourceRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo *CloudControllerApplicationSourceRepository) CopyApplication(sourceAppGUID, targetAppGUID string) error {
	url := fmt.Sprintf("/v2/apps/%s/copy_bits", targetAppGUID)
	body := fmt.Sprintf(`{"source_app_guid":"%s"}`, sourceAppGUID)
	return repo.gateway.CreateResource(repo.config.APIEndpoint(), url, strings.NewReader(body), new(interface{}))
}
