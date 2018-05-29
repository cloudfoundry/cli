package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/actor/versioncheck"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func IsVersionMet(minVersion string) bool {
	version := matchMajorAPIVersion(minVersion)
	ok, err := versioncheck.IsMinimumAPIVersionMet(version, minVersion)
	Expect(err).ToNot(HaveOccurred())

	return ok
}

func matchMajorAPIVersion(minVersion string) string {
	version := getAPIVersionV2()
	if strings.HasPrefix(minVersion, "3") {
		version = getAPIVersionV3()
	}
	return version
}

func SkipIfVersionLessThan(minVersion string) {
	if ignoreAPIVersion() {
		return
	}

	version := matchMajorAPIVersion(minVersion)
	if !IsVersionMet(minVersion) {
		Skip(fmt.Sprintf("minimum version %s not met by API version %s", minVersion, version))
	}
}

func SkipIfVersionAtLeast(maxVersion string) {
	version := matchMajorAPIVersion(maxVersion)

	if IsVersionMet(maxVersion) {
		Skip(fmt.Sprintf("maximum version %s exceeded by API version %s", maxVersion, version))
	}
}

func ignoreAPIVersion() bool {
	ignoreEnv := os.Getenv("CF_INT_IGNORE_API_VERSION_CHECK")
	if ignoreEnv == "" {
		return false
	}

	ignoreBool, err := strconv.ParseBool(ignoreEnv)
	return ignoreBool || err != nil
}

type ccRoot struct {
	Links struct {
		CloudContollerV2 struct {
			Meta struct {
				Version string
			}
		} `json:"cloud_controller_v2"`

		CloudContollerV3 struct {
			Meta struct {
				Version string
			}
		} `json:"cloud_controller_v3"`
	} `json:"links"`
}

func getAPIVersionV2() string {
	return fetchAPIVersion().Links.CloudContollerV2.Meta.Version
}

func getAPIVersionV3() string {
	return fetchAPIVersion().Links.CloudContollerV3.Meta.Version
}

// TODO: Look into caching this
func fetchAPIVersion() ccRoot {
	session := CF("curl", "/")
	Eventually(session).Should(Exit(0))

	var cc ccRoot
	err := json.Unmarshal(session.Out.Contents(), &cc)
	Expect(err).ToNot(HaveOccurred())
	return cc
}
