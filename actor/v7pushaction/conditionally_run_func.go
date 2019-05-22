package v7pushaction

func (actor Actor) ConditionallyRunFunc(condition func(PushPlan) bool, changeFunc ChangeApplicationFunc) ChangeApplicationFunc {
	return func(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
		if condition(pushPlan) {
			return changeFunc(pushPlan, eventStream, progressBar)
		}

		return pushPlan, nil, nil
	}
}
