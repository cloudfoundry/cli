package v7action

func (actor Actor) GetLogCacheEndpoint() string {
	// TODO test this
	 info, _, _, _ := actor.CloudControllerClient.GetInfo()
	 return info.LogCache()
}
