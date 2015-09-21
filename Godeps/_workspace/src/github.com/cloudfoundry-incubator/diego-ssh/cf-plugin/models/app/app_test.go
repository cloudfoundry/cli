package app_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App", func() {
	var (
		fakeCliConnection *fakes.FakeCliConnection
		curler            models.Curler
		af                app.AppFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
	})

	JustBeforeEach(func() {
		af = app.NewAppFactory(fakeCliConnection, curler)
	})

	Describe("Get", func() {
		Context("when CC returns a valid app guid", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
					Expect(args).To(ConsistOf("app", "app1", "--guid"))
					return []string{"app1-guid\n"}, nil
				}
			})

			Context("when an App is returned", func() {
				BeforeEach(func() {
					curler = func(cli plugin.CliConnection, result interface{}, args ...string) error {
						a, ok := result.(*app.CFApp)
						Expect(ok).To(BeTrue())
						a.Metadata.Guid = "app1-guid"
						a.Entity.EnableSSH = true
						a.Entity.Diego = true
						a.Entity.State = "STARTED"
						return nil
					}
				})

				It("returns a populated App model", func() {
					model, err := af.Get("app1")

					Expect(err).NotTo(HaveOccurred())
					Expect(model.Guid).To(Equal("app1-guid"))
					Expect(model.EnableSSH).To(BeTrue())
					Expect(model.Diego).To(BeTrue())
					Expect(model.State).To(Equal("STARTED"))
				})
			})

			Context("when curling the App fails", func() {
				BeforeEach(func() {
					curler = func(cli plugin.CliConnection, result interface{}, args ...string) error {
						return errors.New("not good")
					}
				})

				It("returns an error", func() {
					_, err := af.Get("app1")
					Expect(err).To(MatchError("Failed to acquire app1 info"))
				})
			})
		})

		Context("when the app does not exist", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputReturns(
					[]string{"FAILED", "App app1 is not found"},
					errors.New("Error executing cli core command"),
				)
			})

			It("returns 'App not found' error", func() {
				_, err := af.Get("app1")
				Expect(err).To(MatchError("App app1 is not found"))

				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				args := fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(args).To(ConsistOf("app", "app1", "--guid"))
			})
		})
	})
})
