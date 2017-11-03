package plugin_test

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/plugin"
	"code.cloudfoundry.org/cli/command/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("uninstall-plugin command", func() {
	var (
		cmd        UninstallPluginCommand
		testUI     *ui.UI
		fakeConfig *commandfakes.FakeConfig
		fakeActor  *pluginfakes.FakeUninstallPluginActor
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		somePlugin := configv3.Plugin{
			Name: "some-plugin",
			Version: configv3.PluginVersion{
				Major: 1,
				Minor: 2,
				Build: 3,
			},
		}
		fakeConfig.GetPluginCaseInsensitiveReturns(somePlugin, true)
		fakeActor = new(pluginfakes.FakeUninstallPluginActor)

		cmd = UninstallPluginCommand{
			UI:     testUI,
			Config: fakeConfig,
			Actor:  fakeActor,
		}

		cmd.RequiredArgs.PluginName = "some-plugin"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when the plugin is installed", func() {
		BeforeEach(func() {
			fakeActor.UninstallPluginReturns(nil)
		})

		Context("when no errors are encountered", func() {
			It("uninstalls the plugin and outputs the success message", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("Uninstalling plugin some-plugin..."))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("Plugin some-plugin 1.2.3 successfully uninstalled."))

				Expect(fakeActor.UninstallPluginCallCount()).To(Equal(1))
				_, pluginName := fakeActor.UninstallPluginArgsForCall(0)
				Expect(pluginName).To(Equal("some-plugin"))
			})
		})

		Context("when uninstalling the plugin returns a plugin binary remove failed error", func() {
			var pathError error

			BeforeEach(func() {
				pathError = &os.PathError{
					Op:  "some-op",
					Err: errors.New("some error"),
				}

				fakeActor.UninstallPluginReturns(actionerror.PluginBinaryRemoveFailedError{
					Err: pathError,
				})
			})

			It("returns a PluginBinaryRemoveFailedError", func() {
				Expect(executeErr).To(MatchError(translatableerror.PluginBinaryRemoveFailedError{
					Err: pathError,
				}))
			})
		})

		Context("when uninstalling the plugin returns a plugin execute error", func() {
			var pathError error

			BeforeEach(func() {
				pathError = &os.PathError{
					Op:  "some-op",
					Err: errors.New("some error"),
				}

				fakeActor.UninstallPluginReturns(actionerror.PluginExecuteError{Err: pathError})
			})

			It("returns a PluginBinaryUninstallError", func() {
				Expect(executeErr).To(MatchError(translatableerror.PluginBinaryUninstallError{
					Err: pathError,
				}))
			})
		})

		Context("when uninstalling the plugin encounters any other error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeActor.UninstallPluginReturns(expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})

	Context("when the plugin is not installed", func() {
		BeforeEach(func() {
			fakeActor.UninstallPluginReturns(
				actionerror.PluginNotFoundError{
					PluginName: "some-plugin",
				},
			)
		})

		It("returns a PluginNotFoundError", func() {
			Expect(testUI.Out).To(Say("Uninstalling plugin some-plugin..."))
			Expect(executeErr).To(MatchError(actionerror.PluginNotFoundError{
				PluginName: "some-plugin",
			}))
		})
	})
})
