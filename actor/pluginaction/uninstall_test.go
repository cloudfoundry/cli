package pluginaction_test

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin actor", func() {
	var (
		actor      *Actor
		fakeConfig *pluginactionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		actor = NewActor(fakeConfig, nil)
	})

	Describe("UninstallPlugin", func() {
		var (
			binaryPath            string
			fakePluginUninstaller *pluginactionfakes.FakePluginUninstaller
			pluginHome            string
		)

		BeforeEach(func() {
			var err error
			pluginHome, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			binaryPath = filepath.Join(pluginHome, "banana-faceman")
			err = ioutil.WriteFile(binaryPath, nil, 0600)
			Expect(err).ToNot(HaveOccurred())

			fakePluginUninstaller = new(pluginactionfakes.FakePluginUninstaller)
		})

		AfterEach(func() {
			os.RemoveAll(pluginHome)
		})

		Context("when the plugin does not exist", func() {
			BeforeEach(func() {
				fakeConfig.GetPluginReturns(configv3.Plugin{}, false)
			})

			It("returns a PluginNotFoundError", func() {
				err := actor.UninstallPlugin(fakePluginUninstaller, "some-non-existent-plugin")
				Expect(err).To(MatchError(actionerror.PluginNotFoundError{PluginName: "some-non-existent-plugin"}))
			})
		})

		Context("when the plugin exists", func() {
			BeforeEach(func() {
				fakeConfig.GetPluginReturns(configv3.Plugin{
					Name:     "some-plugin",
					Location: binaryPath,
				}, true)
			})

			Context("when no errors are encountered", func() {
				It("runs the plugin cleanup, deletes the binary and removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeConfig.GetPluginCallCount()).To(Equal(1))
					Expect(fakeConfig.GetPluginArgsForCall(0)).To(Equal("some-plugin"))

					Expect(fakePluginUninstaller.RunCallCount()).To(Equal(1))
					path, command := fakePluginUninstaller.RunArgsForCall(0)
					Expect(path).To(Equal(binaryPath))
					Expect(command).To(Equal("CLI-MESSAGE-UNINSTALL"))

					_, err = os.Stat(binaryPath)
					Expect(os.IsNotExist(err)).To(BeTrue())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
					Expect(fakeConfig.RemovePluginArgsForCall(0)).To(Equal("some-plugin"))

					Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
				})
			})

			Context("when the plugin binary does not exist", func() {
				BeforeEach(func() {
					Expect(os.Remove(binaryPath)).ToNot(HaveOccurred())
				})

				It("removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakePluginUninstaller.RunCallCount()).To(Equal(0))

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
					Expect(fakeConfig.RemovePluginArgsForCall(0)).To(Equal("some-plugin"))

					Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
				})
			})

			Context("when the plugin uninstaller returns an os.PathError", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = &os.PathError{}
					fakePluginUninstaller.RunReturns(expectedErr)
				})

				It("returns a PluginExecuteError, deletes the binary and removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).To(MatchError(actionerror.PluginExecuteError{Err: expectedErr}))

					_, err = os.Stat(binaryPath)
					Expect(os.IsNotExist(err)).To(BeTrue())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
					Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
				})
			})

			Context("when the plugin uninstaller returns an exec.ExitError", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = &exec.ExitError{}
					fakePluginUninstaller.RunReturns(expectedErr)
				})

				It("returns the error, deletes the binary and removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).To(MatchError(actionerror.PluginExecuteError{Err: expectedErr}))

					_, err = os.Stat(binaryPath)
					Expect(os.IsNotExist(err)).To(BeTrue())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
					Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
				})
			})

			Context("when the plugin uninstaller returns any other error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some error")
					fakePluginUninstaller.RunReturns(expectedErr)
				})

				It("returns the error and does not delete the binary or remove the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).To(MatchError(expectedErr))

					_, err = os.Stat(binaryPath)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(0))
				})
			})

			Context("when deleting the plugin binary returns a 'file does not exist' error", func() {
				BeforeEach(func() {
					err := os.Remove(binaryPath)
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not return the error and removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
				})
			})

			Context("when deleting the plugin binary returns a path error", func() {
				BeforeEach(func() {
					err := os.Remove(binaryPath)
					Expect(err).ToNot(HaveOccurred())
					err = os.Mkdir(binaryPath, 0700)
					Expect(err).ToNot(HaveOccurred())
					err = ioutil.WriteFile(filepath.Join(binaryPath, "foooooo"), nil, 0500)
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns the error and removes the plugin config", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					pluginBinaryRemoveErr, ok := err.(actionerror.PluginBinaryRemoveFailedError)
					Expect(ok).To(BeTrue())
					_, isPathError := pluginBinaryRemoveErr.Err.(*os.PathError)
					Expect(isPathError).To(BeTrue())

					Expect(fakeConfig.RemovePluginCallCount()).To(Equal(1))
					Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
				})
			})

			Context("when writing the config returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some plugin config write error")
					fakeConfig.WritePluginConfigReturns(expectedErr)
				})

				It("returns the error", func() {
					err := actor.UninstallPlugin(fakePluginUninstaller, "some-plugin")
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})
	})
})
