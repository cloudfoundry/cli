package v7pushaction

func (actor Actor) IsDropletPathSet(pushPlan PushPlan) bool {
	return pushPlan.DropletPath != ""
}
