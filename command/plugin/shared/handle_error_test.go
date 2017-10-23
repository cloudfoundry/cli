package shared_test

import (
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	. "code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleError", func() {
	err := errors.New("some-error")
	jsonErr := new(json.SyntaxError)
	genericErr := errors.New("some-generic-error")

	DescribeTable("error translations",
		func(passedInErr error, expectedErr error) {
			actualErr := HandleError(passedInErr)
			Expect(actualErr).To(MatchError(expectedErr))
		},

		Entry("json.SyntaxError -> JSONSyntaxError",
			jsonErr,
			translatableerror.JSONSyntaxError{Err: jsonErr},
		),

		Entry("pluginerror.RawHTTPStatusError -> DownloadPluginHTTPError",
			pluginerror.RawHTTPStatusError{Status: "some status"},
			translatableerror.DownloadPluginHTTPError{Message: "some status"},
		),
		Entry("pluginerror.SSLValidationHostnameError -> DownloadPluginHTTPError",
			pluginerror.SSLValidationHostnameError{Message: "some message"},
			translatableerror.DownloadPluginHTTPError{Message: "some message"},
		),
		Entry("pluginerror.UnverifiedServerError -> DownloadPluginHTTPError",
			pluginerror.UnverifiedServerError{URL: "some URL"},
			translatableerror.DownloadPluginHTTPError{Message: "x509: certificate signed by unknown authority"},
		),

		Entry("actionerror.AddPluginRepositoryError -> AddPluginRepositoryError",
			actionerror.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"},
			translatableerror.AddPluginRepositoryError{Name: "some-repo", URL: "some-URL", Message: "404"}),
		Entry("actionerror.GettingPluginRepositoryError -> GettingPluginRepositoryError",
			actionerror.GettingPluginRepositoryError{Name: "some-repo", Message: "404"},
			translatableerror.GettingPluginRepositoryError{Name: "some-repo", Message: "404"}),
		Entry("actionerror.NoCompatibleBinaryError -> NoCompatibleBinaryError",
			actionerror.NoCompatibleBinaryError{},
			translatableerror.NoCompatibleBinaryError{}),
		Entry("actionerror.PluginCommandConflictError -> PluginCommandConflictError",
			actionerror.PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			},
			translatableerror.PluginCommandsConflictError{
				PluginName:     "some-plugin",
				PluginVersion:  "1.1.1",
				CommandNames:   []string{"some-command", "some-other-command"},
				CommandAliases: []string{"sc", "soc"},
			}),
		Entry("actionerror.PluginInvalidError -> PluginInvalidError",
			actionerror.PluginInvalidError{},
			translatableerror.PluginInvalidError{}),
		Entry("actionerror.PluginInvalidError -> PluginInvalidError",
			actionerror.PluginInvalidError{Err: genericErr},
			translatableerror.PluginInvalidError{Err: genericErr}),
		Entry("actionerror.PluginNotFoundError -> PluginNotFoundError",
			actionerror.PluginNotFoundError{PluginName: "some-plugin"},
			translatableerror.PluginNotFoundError{PluginName: "some-plugin"}),
		Entry("actionerror.RepositoryNameTakenError -> RepositoryNameTakenError",
			actionerror.RepositoryNameTakenError{Name: "some-repo"},
			translatableerror.RepositoryNameTakenError{Name: "some-repo"}),
		Entry("actionerror.RepositoryNotRegisteredError -> RepositoryNotRegisteredError",
			actionerror.RepositoryNotRegisteredError{Name: "some-repo"},
			translatableerror.RepositoryNotRegisteredError{Name: "some-repo"}),

		Entry("default case -> original error",
			err,
			err),
	)

	It("returns nil for a common.PluginInstallationCancelled error", func() {
		err := HandleError(PluginInstallationCancelled{})
		Expect(err).To(BeNil())
	})

	It("returns nil for a nil error", func() {
		nilErr := HandleError(nil)
		Expect(nilErr).To(BeNil())
	})
})
