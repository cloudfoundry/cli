package readonly

import (
	"fmt"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var variableNames map[string]string

func init() {
	variableNames = make(map[string]string)
}

/** This method sets up all of the read only resources needed during read only tests
GLOBAL SETUP:
1) SetupOrgs => create an empty org and an org
2) SetupStacks => create a new bogus stack
3) SetupBuildpacks => create a new bogus buildpack (uses curl to avoid upload)
4) SetupFeatureFlags => may be a no-op Is a noop X


ORG LEVEL SETUP:
1) SetupSpaces => create an empty space and a space
2) SetupDomains => create a shared and private domain

SPACE LEVEL SETUP:
1) SetupApps => push any apps we need (creates 3? apps maybe one with just create app, one thats running, and one thats stopped??)
2) SetupRoutes (creates a route mapped to an app as well as one that is not mapped to anything)

Each will create a random name and add it to a map at the location of a well defined key

EmptyOrg
NonEmptyOrg
EmptySpace
NonEmptySpace
SharedDomain
PrivateDomain
SharedPrivateDomain
Stack
Buildpack
RunningApp
StoppedApp
AppSkeleton
RouteWithDestinations
RouteWithNoDestinations

**/
func SetupReadOnlySuite() {
	// login as admin

	// Global level setup - This will set up foundation wide resources
	SetupOrgs()
	SetupStacks()
	SetupBuildpacks()
	// FeatureFlags are global resource but no setup is necessary, add here if that changes

	// Org level setup - This sets up resources conatined within an org
	SetupSpaces()
	SetupDomains()

	// Space level setup - this sets up resources at a space level
	SetupApps()
	SetupRoutes()
}

func SetupOrgs() {
	emptyOrgName := helpers.NewOrgName()
	Eventually(helpers.CF("create-org"), emptyOrgName).Should(Exit(0))
	variableNames["emptyOrg"] = emptyOrgName

	nonEmptyOrgName := helpers.NewOrgName()
	Eventually(helpers.CF("create-org"), nonEmptyOrgName).Should(Exit(0))
	variableNames["nonEmptyOrg"] = nonEmptyOrgName
}

func SetupStacks() {
	stackName := helpers.NewStackName()
	Eventually(helpers.CF("curl", "-X", "POST", "-d", fmt.Sprintf(`{"name": %s}`, stackName), "/v3/stacks"))
	variableNames["stackName"] = stackName
}

func SetupBuildpacks() {
	buildpackName := helpers.NewBuildpackName()
	Eventually(helpers.CF("curl", "-X", "POST", "-d", fmt.Sprintf(`{"name": %s}`, buildpackName), "/v3/buildpcks"))
	variableNames["buildpackName"] = buildpackName

}

func SetupSpaces() {
	emptySpaceName := helpers.NewSpaceName()
	Eventually(helpers.CF("create-space"), emptySpaceName, "-o", variableNames["nonEmptyOrgName"]).Should(Exit(0))
	variableNames["emptySpace"] = emptySpaceName

	nonEmptySpaceName := helpers.NewSpaceName()
	Eventually(helpers.CF("create-space"), nonEmptySpaceName, "-o", variableNames["nonEmptyOrgName"]).Should(Exit(0))
	variableNames["nonEmptySpace"] = nonEmptySpaceName
}

func SetupDomains() {

}

func SetupApps() {

}

func SetupRoutes() {

}
