package isolated

import (
	"testing"

	"code.cloudfoundry.org/cli/integration/helpers/commonisolated"
)

const (
	RealIsolationSegment = commonisolated.RealIsolationSegment
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation bool
	ReadOnlyOrg       string
	ReadOnlySpace     string

	// Per Test Level
	homeDir string
)

func TestIsolated(t *testing.T) {
	commonisolated.CommonTestIsolated(t)
}

var _ = commonisolated.CommonGinkgoSetup(
	"summary_isi.txt",
	&apiURL,
	&skipSSLValidation,
	&ReadOnlyOrg,
	&ReadOnlySpace,
	&homeDir,
)
