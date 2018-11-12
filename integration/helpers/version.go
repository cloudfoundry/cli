package helpers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

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

type UAAVersion struct {
	App struct {
		Version string `json:"version"`
	} `json:"app"`
}

func (v UAAVersion) Version() string {
	return v.App.Version
}

func IsUAAVersionAtLeast(minVersion string) bool {
	info := fetchAPIVersion()
	uaaUrl := fmt.Sprintf("%s/info", info.Links.UAA.Href)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: len(skipSSLValidation()) > 0},
	}
	req, _ := http.NewRequest("GET", uaaUrl, nil)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	Expect(err).ToNot(HaveOccurred())

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		Expect(err2).ToNot(HaveOccurred())

		version := &UAAVersion{}

		err3 := json.Unmarshal(bodyBytes, &version)
		Expect(err3).ToNot(HaveOccurred())
		currentUaaVersion := version.Version()
		ok, err := versioncheck.IsMinimumAPIVersionMet(currentUaaVersion, minVersion)
		Expect(err).ToNot(HaveOccurred())
		return ok
	}
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	return false
}

func SkipIfUAAVersionLessThan(version string) {
	if !IsUAAVersionAtLeast(version) {
		Skip(fmt.Sprintf("Test requires UAA version at least %s", version))
	}
}

func SkipIfUAAVersionAtLeast(version string) {
	if IsUAAVersionAtLeast(version) {
		Skip(fmt.Sprintf("Test requires UAA version less than %s", version))
	}
}

func matchMajorAPIVersion(minVersion string) string {
	version := GetAPIVersionV2()
	if strings.HasPrefix(minVersion, "3") {
		version = getAPIVersionV3()
	}
	return version
}

func GetAPIVersionV2() string {
	return fetchAPIVersion().Links.CloudContollerV2.Meta.Version
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

		UAA struct {
			Href string `json:"href"`
		} `json:"uaa"`
	} `json:"links"`
}

var cacheLock sync.Mutex
var CcRootCache *ccRoot

func fetchAPIVersion() ccRoot {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if CcRootCache == nil {
		session := CF("curl", "/")
		Eventually(session).Should(Exit(0))
		var result ccRoot
		err := json.Unmarshal(session.Out.Contents(), &result)
		Expect(err).ToNot(HaveOccurred())
		CcRootCache = &result
	}
	return *CcRootCache
}

func getAPIVersionV3() string {
	return fetchAPIVersion().Links.CloudContollerV3.Meta.Version
}

func SkipIfNoRoutingAPI() {
	// TODO: #161159794 remove this function and check a nicer error message when available
	var response struct {
		RoutingEndpoint string `json:"routing_endpoint"`
	}
	Curl(&response, "/v2/info")

	if response.RoutingEndpoint == "" {
		Skip("Test requires routing endpoint on /v2/info")
	}
}
