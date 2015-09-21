package cmd_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/cmd"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space/space_fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisallowSSH", func() {
	var fakeSpaceFactory *space_fakes.FakeSpaceFactory
	var mySpace space.Space

	BeforeEach(func() {
		fakeSpaceFactory = &space_fakes.FakeSpaceFactory{}
		mySpace = space.Space{Guid: "myguid"}
	})

	Context("validation", func() {
		It("requires an space name", func() {
			err := cmd.DisallowSSH([]string{"disallow-space-ssh"}, fakeSpaceFactory)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.DisallowSSHUsage))
		})

		It("validates the command name", func() {
			err := cmd.DisallowSSH([]string{"bogus", "space"}, fakeSpaceFactory)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.DisallowSSHUsage))
		})
	})

	It("disallows SSH on an space endpoint", func() {
		fakeSpaceFactory.GetReturns(mySpace, nil)

		err := cmd.DisallowSSH([]string{"disallow-space-ssh", "myspace"}, fakeSpaceFactory)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeSpaceFactory.GetCallCount()).To(Equal(1))
		Expect(fakeSpaceFactory.GetArgsForCall(0)).To(Equal("myspace"))

		Expect(fakeSpaceFactory.SetBoolCallCount()).To(Equal(1))
		aSpace, key, val := fakeSpaceFactory.SetBoolArgsForCall(0)
		Expect(aSpace).To(Equal(mySpace))
		Expect(key).To(Equal("allow_ssh"))
		Expect(val).To(BeFalse())
	})

	Context("when retrieving the Space fails", func() {
		BeforeEach(func() {
			fakeSpaceFactory.GetReturns(space.Space{}, errors.New("get failed"))
		})

		It("returns an err", func() {
			err := cmd.DisallowSSH([]string{"disallow-space-ssh", "myspace"}, fakeSpaceFactory)
			Expect(err).To(MatchError("get failed"))
			Expect(fakeSpaceFactory.GetCallCount()).To(Equal(1))
			Expect(fakeSpaceFactory.SetBoolCallCount()).To(Equal(0))
		})
	})

	Context("when setting the value fails", func() {
		BeforeEach(func() {
			fakeSpaceFactory.GetReturns(mySpace, nil)
			fakeSpaceFactory.SetBoolReturns(errors.New("set failed"))
		})

		It("returns an err", func() {
			err := cmd.DisallowSSH([]string{"disallow-space-ssh", "myspace"}, fakeSpaceFactory)
			Expect(err).To(MatchError("set failed"))
			Expect(fakeSpaceFactory.GetCallCount()).To(Equal(1))
			Expect(fakeSpaceFactory.SetBoolCallCount()).To(Equal(1))
		})
	})
})
