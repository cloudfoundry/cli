package v7action

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) CreateSecurityGroup(name, filePath string) (Warnings, error) {
	allWarnings := Warnings{}
	bytes, err := parsePath(filePath)
	if err != nil {
		return allWarnings, err
	}

	var rules []resources.Rule
	err = json.Unmarshal(bytes, &rules)
	if err != nil {
		return allWarnings, err
	}

	securityGroup := resources.SecurityGroup{Name: name, Rules: rules}

	_, warnings, err := actor.CloudControllerClient.CreateSecurityGroup(securityGroup)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func parsePath(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
