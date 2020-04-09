package servicebrokerstub

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
)

const (
	appNamePrefix      = "hydrabroker"
	appOrg             = "fakeservicebroker"
	appSpace           = "integration"
	defaultMemoryLimit = "32M"
	pathToApp          = "../../assets/hydrabroker"
)

func ensureAppIsDeployed() {
	if !appResponds() {
		ensureAppIsPushed()
		Eventually(appResponds).Should(BeTrue())
	}
}

func appResponds() bool {
	resp, err := http.Get(appURL("/config"))
	Expect(err).ToNot(HaveOccurred())
	defer resp.Body.Close()
	b := resp.StatusCode == http.StatusOK
	return b
}

func ensureAppIsPushed() {
	appExists := func() bool {
		session := helpers.CF("app", "--guid", appName())
		session.Wait()
		return session.ExitCode() == 0
	}

	pushApp := func() bool {
		session := helpers.CF(
			"push", appName(),
			"-p", pathToApp,
			"-m", defaultMemoryLimit,
		)
		session.Wait()
		return session.ExitCode() == 0
	}

	cleanupAppsFromPreviousRuns := func() {
		session := helpers.CF("apps")
		session.Wait()

		if session.ExitCode() == 0 {
			matchingApps := regexp.MustCompile(fmt.Sprintf(`%s-\d+`, appNamePrefix)).
				FindAllString(string(session.Out.Contents()), -1)

			for _, app := range matchingApps {
				if app != appName() {
					session := helpers.CF("delete", app, "-f")
					session.Wait()
				}
			}
		}
	}

	helpers.CreateOrgAndSpaceUnlessExists(appOrg, appSpace)
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.TargetOrgAndSpace(appOrg, appSpace)

		cleanupAppsFromPreviousRuns()

		ok := false
		for attempts := 0; attempts < 5 && !ok; attempts++ {
			ok = appExists()
			if !ok {
				ok = pushApp()
			}
			if !ok {
				time.Sleep(5 * time.Second)
			}
		}

		Expect(ok).To(BeTrue(), "Failed to push app")
	})
}

func appURL(paths ...string) string {
	return fmt.Sprintf("http://%s.%s%s", appName(), helpers.DefaultSharedDomain(), strings.Join(paths, ""))
}

func appName() string {
	id := ginkgo.GinkgoRandomSeed()
	if len(os.Getenv("REUSE_SERVICE_BROKER_APP")) > 0 {
		id = 0
	}
	return fmt.Sprintf("%s-%010d", appNamePrefix, id)
}
