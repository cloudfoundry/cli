package v7action

import (
	"code.cloudfoundry.org/cli/types"
)

func (actor *Actor) GetApplicationLabels(appName string, spaceGUID string) (map[string]types.NullString, Warnings, error) {
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	return app.Metadata.Labels, warnings, err
}

func (actor *Actor) UpdateApplicationLabelsByApplicationName(appName string, spaceGUID string, labels map[string]types.NullString) (Warnings, error) {
	app, appWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return appWarnings, err
	}
	app.Metadata.Labels = labels
	_, updateWarnings, err := actor.UpdateApplication(app)
	return append(appWarnings, updateWarnings...), err
}
