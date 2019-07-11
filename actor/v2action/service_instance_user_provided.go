package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// UserProvidedServiceInstance represents an instance of a user-provided service.
type UserProvidedServiceInstance ccv2.UserProvidedServiceInstance

func (u UserProvidedServiceInstance) WithTags(tags []string) UserProvidedServiceInstance {
	if tags == nil {
		tags = []string{}
	}

	u.Tags = &tags
	return u
}

func (u UserProvidedServiceInstance) WithSyslogDrainURL(url string) UserProvidedServiceInstance {
	u.SyslogDrainURL = &url
	return u
}

func (u UserProvidedServiceInstance) WithRouteServiceURL(url string) UserProvidedServiceInstance {
	u.RouteServiceURL = &url
	return u
}

func (u UserProvidedServiceInstance) WithCredentials(creds map[string]interface{}) UserProvidedServiceInstance {
	if creds == nil {
		creds = make(map[string]interface{})
	}

	u.Credentials = creds
	return u
}

// UpdateUserProvidedServiceInstance requests that the service instance be updated with the specified data
func (actor Actor) UpdateUserProvidedServiceInstance(guid string, instance UserProvidedServiceInstance) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.UpdateUserProvidedServiceInstance(guid, ccv2.UserProvidedServiceInstance(instance))
	return Warnings(warnings), err
}
