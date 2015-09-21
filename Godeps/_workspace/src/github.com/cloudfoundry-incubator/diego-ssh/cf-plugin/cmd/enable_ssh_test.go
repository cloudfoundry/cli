package cmd_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/cmd"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/app/app_fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnableSsh", func() {
	var fakeAppFactory *app_fakes.FakeAppFactory
	var myApp app.App

	BeforeEach(func() {
		fakeAppFactory = &app_fakes.FakeAppFactory{}
		myApp = app.App{Guid: "myguid"}
	})

	Context("validation", func() {
		It("requires an application name", func() {
			err := cmd.EnableSSH([]string{"enable-ssh"}, fakeAppFactory)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.EnableSSHUsage))
		})

		It("validates the command name", func() {
			err := cmd.EnableSSH([]string{"enable-ss", "app"}, fakeAppFactory)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.EnableSSHUsage))
		})
	})

	It("enables SSH on an app endpoint", func() {
		fakeAppFactory.GetReturns(myApp, nil)

		err := cmd.EnableSSH([]string{"enable-ssh", "myapp"}, fakeAppFactory)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
		Expect(fakeAppFactory.GetArgsForCall(0)).To(Equal("myapp"))

		Expect(fakeAppFactory.SetBoolCallCount()).To(Equal(1))
		anApp, key, val := fakeAppFactory.SetBoolArgsForCall(0)
		Expect(anApp).To(Equal(myApp))
		Expect(key).To(Equal("enable_ssh"))
		Expect(val).To(BeTrue())
	})

	Context("when retrieving the App fails", func() {
		BeforeEach(func() {
			fakeAppFactory.GetReturns(app.App{}, errors.New("get failed"))
		})

		It("returns an err", func() {
			err := cmd.EnableSSH([]string{"enable-ssh", "myapp"}, fakeAppFactory)
			Expect(err).To(MatchError("get failed"))
			Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
			Expect(fakeAppFactory.SetBoolCallCount()).To(Equal(0))
		})
	})

	Context("when setting the value fails", func() {
		BeforeEach(func() {
			fakeAppFactory.GetReturns(myApp, nil)
			fakeAppFactory.SetBoolReturns(errors.New("set failed"))
		})

		It("returns an err", func() {
			err := cmd.EnableSSH([]string{"enable-ssh", "myapp"}, fakeAppFactory)
			Expect(err).To(MatchError("set failed"))
			Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
			Expect(fakeAppFactory.SetBoolCallCount()).To(Equal(1))
		})
	})
})
