package copy_application_source

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
)

type CopyApplicationSourceRepository interface {
	CopyApplication(sourceAppGuid, targetAppGuid string) error
}

type CloudControllerApplicationSourceRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerCopyApplicationSourceRepository(config core_config.Reader, gateway net.Gateway) *CloudControllerApplicationSourceRepository {
	return &CloudControllerApplicationSourceRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo *CloudControllerApplicationSourceRepository) CopyApplication(sourceAppGuid, targetAppGuid string) error {
	url := fmt.Sprintf("/v2/apps/%s/copy_bits", targetAppGuid)
	body := fmt.Sprintf(`{"source_app_guid":"%s"}`, sourceAppGuid)
	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), url, strings.NewReader(body), new(interface{}))
}
