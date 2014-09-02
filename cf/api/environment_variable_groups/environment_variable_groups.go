package environment_variable_groups

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type EnvironmentVariableGroupsRepository interface {
	ListRunning() (variables []models.EnvironmentVariable, apiErr error)
}

type CloudControllerEnvironmentVariableGroupsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerEnvironmentVariableGroupsRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerEnvironmentVariableGroupsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerEnvironmentVariableGroupsRepository) ListRunning() (variables []models.EnvironmentVariable, apiErr error) {
	var raw_response interface{}
	url := fmt.Sprintf("%s/v2/config/environment_variable_groups/running", repo.config.ApiEndpoint())
	apiErr = repo.gateway.GetResource(url, &raw_response)
	if apiErr != nil {
		return
	}

	variables, err := repo.marshalToEnvironmentVariables(raw_response)
	if err != nil {
		return nil, err
	}

	return variables, nil
}

func (repo CloudControllerEnvironmentVariableGroupsRepository) marshalToEnvironmentVariables(raw_response interface{}) ([]models.EnvironmentVariable, error) {
	var variables []models.EnvironmentVariable
	for key, value := range raw_response.(map[string]interface{}) {
		stringvalue, err := repo.convertValueToString(value)
		if err != nil {
			return nil, err
		}
		variable := models.EnvironmentVariable{Name: key, Value: stringvalue}
		variables = append(variables, variable)
	}
	return variables, nil
}

func (repo CloudControllerEnvironmentVariableGroupsRepository) convertValueToString(value interface{}) (string, error) {
	stringvalue, ok := value.(string)
	if !ok {
		floatvalue, ok := value.(float64)
		if !ok {
			return "", errors.New(fmt.Sprintf("Attempted to read environment variable value of unknown type: %#v", value))
		}
		stringvalue = fmt.Sprintf("%d", int(floatvalue))
	}
	return stringvalue, nil
}
