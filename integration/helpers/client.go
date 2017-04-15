package helpers

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/gomega"
)

func CreateV2Client(homeDir string) *ccv2.Client {
	file, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
	Expect(err).ToNot(HaveOccurred())

	var config configv3.Config
	err = json.Unmarshal(file, &config.ConfigFile)
	Expect(err).ToNot(HaveOccurred())

	ccWrappers := []ccv2.ConnectionWrapper{}
	authWrapper := ccWrapper.NewUAAAuthentication(nil, &config)
	ccWrappers = append(ccWrappers, authWrapper)
	ccWrappers = append(ccWrappers, ccWrapper.NewRetryRequest(2))

	ccClient := ccv2.NewClient(ccv2.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappers,
	})

	_, err = ccClient.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	Expect(err).ToNot(HaveOccurred())

	uaaClient := uaa.NewClient(uaa.Config{
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		ClientID:          config.UAAOAuthClient(),
		ClientSecret:      config.UAAOAuthClientSecret(),
		DialTimeout:       config.DialTimeout(),
		SkipSSLValidation: config.SkipSSLValidation(),
		URL:               ccClient.TokenEndpoint(),
	})

	uaaClient.WrapConnection(uaaWrapper.NewUAAAuthentication(uaaClient, &config))
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(2))
	authWrapper.SetClient(uaaClient)

	return ccClient
}
