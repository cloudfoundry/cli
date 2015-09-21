package space_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models"
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models/space"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Space", func() {
	var (
		fakeCliConnection *fakes.FakeCliConnection
		curler            models.Curler
		sf                space.SpaceFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}

	})

	JustBeforeEach(func() {
		sf = space.NewSpaceFactory(fakeCliConnection, curler)
	})

	Describe("Get", func() {
		Context("when CC returns a valid space guid", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
					Expect(args).To(ConsistOf("space", "space1", "--guid"))
					return []string{"space1-guid\n"}, nil
				}
			})

			Context("when a Space is returned", func() {
				BeforeEach(func() {
					curler = func(cli plugin.CliConnection, result interface{}, args ...string) error {
						s, ok := result.(*space.CFSpace)
						Expect(ok).To(BeTrue())
						s.Metadata.Guid = "space1-guid"
						s.Entity.AllowSSH = true
						return nil
					}
				})

				It("returns a populated Space model", func() {
					model, err := sf.Get("space1")

					Expect(err).NotTo(HaveOccurred())
					Expect(model.Guid).To(Equal("space1-guid"))
					Expect(model.AllowSSH).To(BeTrue())
				})
			})

			Context("when curling the Space fails", func() {
				BeforeEach(func() {
					curler = func(cli plugin.CliConnection, result interface{}, args ...string) error {
						return errors.New("not good")
					}
				})

				It("returns an error", func() {
					_, err := sf.Get("space1")
					Expect(err).To(MatchError("Failed to acquire space1 info"))
				})
			})
		})

		Context("when the space does not exist", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputReturns(
					[]string{"FAILED", "Space space1 is not found"},
					errors.New("Error executing cli core command"),
				)
			})

			It("returns 'Space not found' error", func() {
				_, err := sf.Get("space1")
				Expect(err).To(MatchError("Space space1 is not found"))

				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
				args := fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)
				Expect(args).To(ConsistOf("space", "space1", "--guid"))
			})
		})
	})
})
