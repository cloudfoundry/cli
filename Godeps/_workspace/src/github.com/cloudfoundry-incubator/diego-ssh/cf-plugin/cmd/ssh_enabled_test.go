package cmd_test

import (
	"bytes"
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/cmd"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app/app_fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHEnabled", func() {
	var fakeAppFactory *app_fakes.FakeAppFactory
	var myApp app.App

	BeforeEach(func() {
		fakeAppFactory = &app_fakes.FakeAppFactory{}
		myApp = app.App{Guid: "myguid"}
	})

	Context("validation", func() {
		It("requires an application name", func() {
			err := cmd.SSHEnabled([]string{"ssh-enabled"}, fakeAppFactory, nil)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.SSHEnabledUsage))
		})

		It("validates the command name", func() {
			err := cmd.SSHEnabled([]string{"ssh-enable", "app"}, fakeAppFactory, nil)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.SSHEnabledUsage))
		})
	})

	It("returns the value", func() {
		myApp.EnableSSH = true
		fakeAppFactory.GetReturns(myApp, nil)
		writer := bytes.NewBuffer(nil)

		err := cmd.SSHEnabled([]string{"ssh-enabled", "myapp"}, fakeAppFactory, writer)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
		Expect(fakeAppFactory.GetArgsForCall(0)).To(Equal("myapp"))
		Expect(writer.String()).To(Equal("true"))
	})

	Context("when retrieving the App fails", func() {
		BeforeEach(func() {
			fakeAppFactory.GetReturns(app.App{}, errors.New("get failed"))
		})

		It("returns an err", func() {
			err := cmd.SSHEnabled([]string{"ssh-enabled", "myapp"}, fakeAppFactory, nil)
			Expect(err).To(MatchError("get failed"))
			Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
		})
	})
})
