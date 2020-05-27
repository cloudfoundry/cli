package v7action

func (actor Actor) GetLogCacheEndpoint() (string, Warnings, error) {
	info, _, warnings, err := actor.CloudControllerClient.GetInfo()
	if err != nil {
		return "", Warnings(warnings), err
	}
	return info.LogCache(), Warnings(warnings), nil
}
