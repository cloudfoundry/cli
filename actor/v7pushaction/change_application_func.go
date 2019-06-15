package v7pushaction

// ChangeApplicationFunc is a function that is used by Actualize to setup application for staging, droplet creation, etc.
type ChangeApplicationFunc func(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error)
