package cmd_test

import (
	"bytes"
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/cmd"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space/space_fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHAllowed", func() {
	var fakeSpaceFactory *space_fakes.FakeSpaceFactory
	var mySpace space.Space

	BeforeEach(func() {
		fakeSpaceFactory = &space_fakes.FakeSpaceFactory{}
		mySpace = space.Space{Guid: "myguid"}
	})

	Context("validation", func() {
		It("requires an spacelication name", func() {
			err := cmd.SSHAllowed([]string{"space-ssh-allowed"}, fakeSpaceFactory, nil)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.SSHAllowedUsage))
		})

		It("validates the command name", func() {
			err := cmd.SSHAllowed([]string{"bogus", "space"}, fakeSpaceFactory, nil)

			Expect(err).To(MatchError("Invalid usage\n" + cmd.SSHAllowedUsage))
		})
	})

	It("returns the value", func() {
		mySpace.AllowSSH = true
		fakeSpaceFactory.GetReturns(mySpace, nil)
		writer := bytes.NewBuffer(nil)

		err := cmd.SSHAllowed([]string{"space-ssh-allowed", "myspace"}, fakeSpaceFactory, writer)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeSpaceFactory.GetCallCount()).To(Equal(1))
		Expect(fakeSpaceFactory.GetArgsForCall(0)).To(Equal("myspace"))
		Expect(writer.String()).To(Equal("true"))
	})

	Context("when retrieving the Space fails", func() {
		BeforeEach(func() {
			fakeSpaceFactory.GetReturns(space.Space{}, errors.New("get failed"))
		})

		It("returns an err", func() {
			err := cmd.SSHAllowed([]string{"space-ssh-allowed", "myspace"}, fakeSpaceFactory, nil)
			Expect(err).To(MatchError("get failed"))
			Expect(fakeSpaceFactory.GetCallCount()).To(Equal(1))
		})
	})
})
