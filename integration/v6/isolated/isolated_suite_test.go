package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers/commonisolated"
	"testing"
)

const (
	RealIsolationSegment = commonisolated.RealIsolationSegment
	DockerImage          = commonisolated.DockerImage
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation bool
	ReadOnlyOrg       string
	ReadOnlySpace     string

	// Per test
	homeDir string
)

func TestIsolated(t *testing.T) {
	commonisolated.CommonTestIsolated(t)
}

var _ = commonisolated.CommonGinkgoSetup(
	"summary_ivi.txt",
	&apiURL,
	&skipSSLValidation,
	&ReadOnlyOrg,
	&ReadOnlySpace,
	&homeDir,
)
