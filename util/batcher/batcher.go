package batcher

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

const BatchSize = 50

type callback func(guids []string) (ccv3.Warnings, error)

func RequestByGUID(guids []string, cb callback) (ccv3.Warnings, error) {
	var allWarnings ccv3.Warnings

	for len(guids) > 0 {
		remaining := len(guids)
		if remaining > BatchSize {
			remaining = BatchSize
		}

		batch := guids[:remaining]
		guids = guids[remaining:]

		warnings, err := cb(batch)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}
	}

	return allWarnings, nil
}
