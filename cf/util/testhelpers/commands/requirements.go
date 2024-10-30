package commands

import "code.cloudfoundry.org/cli/v7/cf/requirements"

func RunRequirements(reqs []requirements.Requirement) error {
	var retErr error

	for _, req := range reqs {
		if err := req.Execute(); err != nil {
			retErr = err
			break
		}
	}

	return retErr
}
