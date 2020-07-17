package util

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type funcWithWarningsAndError = func() (ccv3.Warnings, error)

type Railway struct {
	tracks []funcWithWarningsAndError
}

func StartRailway() Railway {
	return Railway{}
}

func (railway Railway) Track(track funcWithWarningsAndError) Railway {
	railway.tracks = append(railway.tracks, track)
	return railway
}

func (railway Railway) Execute() (ccv3.Warnings, error) {
	var warnings ccv3.Warnings

	for _, track := range railway.tracks {
		trackWarnings, err := track()
		warnings = append(warnings, trackWarnings...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}
