package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type Buildpack ccv3.Buildpack

func (actor Actor) GetBuildpacks() ([]Buildpack, Warnings, error) {
	ccv3Buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{ccv3.PositionOrder},
	})

	var buildpacks []Buildpack
	for _, buildpack := range ccv3Buildpacks {
		buildpacks = append(buildpacks, Buildpack(buildpack))
	}

	return buildpacks, Warnings(warnings), err
}

func (actor Actor) CreateBuildpack(buildpack Buildpack) (Buildpack, Warnings, error) {
	ccv3Buildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(
		ccv3.Buildpack(buildpack),
	)

	return Buildpack(ccv3Buildpack), Warnings(warnings), err
}
