package isolated

import (
	"flag"
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/integration/helpers/commonisolated"
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation bool
	ReadOnlyOrg       string
	ReadOnlySpace     string

	// Per Test Level
	homeDir string
	myFlag  string
)

func init() {
	flag.StringVar(&myFlag, "myFlag", "defaultvalue", "myFlag is used to control my behavior")
}
func TestIsolated(t *testing.T) {
	commonisolated.CommonTestIsolated(t)
	// flag.StringVar(&myFlag, "myFlag", "defaultvalue", "myFlag is used to control my behavior")
	fmt.Println("myFlag =========================", myFlag)
}

var _ = commonisolated.CommonGinkgoSetup(
	"summary_isi.txt",
	&apiURL,
	&skipSSLValidation,
	&ReadOnlyOrg,
	&ReadOnlySpace,
	&homeDir,
)
