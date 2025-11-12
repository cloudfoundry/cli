package railway

import "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"

type funcWithWarningsAndError = func() (ccv3.Warnings, error)

func Sequentially(tracks ...funcWithWarningsAndError) (ccv3.Warnings, error) {
	var warnings ccv3.Warnings

	for _, track := range tracks {
		trackWarnings, err := track()
		warnings = append(warnings, trackWarnings...)
		if err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}
