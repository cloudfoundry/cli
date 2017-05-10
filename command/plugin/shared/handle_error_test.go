package shared_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	. "code.cloudfoundry.org/cli/command/plugin/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleError", func() {
	err := errors.New("some-error")

	DescribeTable("error translations",
		func(passedInErr error, expectedErr error) {
			actualErr := HandleError(passedInErr)
			Expect(actualErr).To(MatchError(expectedErr))
		},

		Entry("pluginerror.RawHTTPStatusError -> DownloadPluginHTTPError",
			pluginerror.RawHTTPStatusError{Status: "some status"},
			DownloadPluginHTTPError{Message: "some status"},
		),
		Entry("pluginerror.SSLValidationHostnameError -> DownloadPluginHTTPError",
			pluginerror.SSLValidationHostnameError{Message: "some message"},
			DownloadPluginHTTPError{Message: "Hostname does not match SSL Certificate (some message)"},
		),
		Entry("pluginerror.UnverifiedServerError -> DownloadPluginHTTPError",
			pluginerror.UnverifiedServerError{URL: "some URL"},
			DownloadPluginHTTPError{Message: "x509: certificate signed by unknown authority"},
		),

		Entry("pluginaction.AddPluginRepositoryError -> AddPluginRepositoryError",
			pluginaction.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"},
			AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"}),
		Entry("pluginaction.GettingPluginRepositoryError -> GettingPluginRepositoryError",
			pluginaction.GettingPluginRepositoryError{Name: "some-repo", Message: "404"},
			GettingPluginRepositoryError{Name: "some-repo", Message: "404"}),
		Entry("pluginaction.PluginCommandConflictError -> PluginCommandConflictError",
			pluginaction.PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			},
			PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			}),
		Entry("pluginaction.PluginInvalidError -> PluginInvalidError",
			pluginaction.PluginInvalidError{},
			PluginInvalidError{}),
		Entry("pluginaction.PluginNotFoundError -> PluginNotFoundError",
			pluginaction.PluginNotFoundError{Name: "some-plugin"},
			PluginNotFoundError{Name: "some-plugin"}),
		Entry("pluginaction.RepositoryNameTakenError -> RepositoryNameTakenError",
			pluginaction.RepositoryNameTakenError{Name: "some-repo"},
			RepositoryNameTakenError{Name: "some-repo"}),

		Entry("default case -> original error",
			err,
			err),
	)

	It("returns nil for a nil error", func() {
		nilErr := HandleError(nil)
		Expect(nilErr).To(BeNil())
	})
})
