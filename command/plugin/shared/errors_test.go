package shared_test

import (
	"bytes"
	"text/template"

	. "code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Translatable Errors", func() {
	translateFunc := func(s string, vars ...interface{}) string {
		formattedTemplate, err := template.New("test template").Parse(s)
		Expect(err).ToNot(HaveOccurred())
		formattedTemplate.Option("missingkey=error")

		var buffer bytes.Buffer
		if len(vars) > 0 {
			err = formattedTemplate.Execute(&buffer, vars[0])
			Expect(err).ToNot(HaveOccurred())

			return buffer.String()
		} else {
			return s
		}
	}

	DescribeTable("translates error",
		func(e error) {
			err, ok := e.(ui.TranslatableError)
			Expect(ok).To(BeTrue())
			err.Translate(translateFunc)
		},

		Entry("PluginNotFoundError", PluginNotFoundError{}),
		Entry("NoPluginRepositoriesError", NoPluginRepositoriesError{}),
		Entry("RepositoryNameTakenError", RepositoryNameTakenError{}),
		Entry("AddPluginRepositoryError", AddPluginRepositoryError{}),
		Entry("GettingPluginRepositoryError", GettingPluginRepositoryError{}),
		Entry("FileNotFoundError", FileNotFoundError{}),
		Entry("PluginInvalidError", PluginInvalidError{}),
		Entry("PluginCommandsConflictError", PluginCommandsConflictError{}),
		Entry("PluginAlreadyInstalledError", PluginAlreadyInstalledError{}),
		Entry("DownloadPluginHTTPError", DownloadPluginHTTPError{}),
	)
})

var _ = Describe("PluginNotFoundError", func() {
	var err PluginNotFoundError

	Context("when the repository is specified", func() {
		BeforeEach(func() {
			err = PluginNotFoundError{PluginName: "some-plugin", RepositoryName: "some-repo"}
		})

		It("errors with the plugin name and with the repository", func() {
			Expect(err.Error()).To(Equal("Plugin {{.PluginName}} not found in repository {{.RepositoryName}}"))
		})
	})

	Context("when the repository is not specified", func() {
		BeforeEach(func() {
			err = PluginNotFoundError{PluginName: "some-plugin"}
		})

		It("errors with the plugin name and without the repository", func() {
			Expect(err.Error()).To(Equal("Plugin {{.PluginName}} does not exist."))
		})
	})
})
