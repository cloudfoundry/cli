package terminal_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/io_helpers"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/terminal"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"strings"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("UI", func() {
	Describe("Printing message to stdout with Say", func() {
		It("prints strings", func() {
			io_helpers.SimulateStdin("", func(reader io.Reader) {
				output := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello")
				})

				Expect("Hello").To(Equal(strings.Join(output, "")))
			})
		})

		It("prints formatted strings", func() {
			io_helpers.SimulateStdin("", func(reader io.Reader) {
				output := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					ui.Say("Hello %s", "World!")
				})

				Expect("Hello World!").To(Equal(strings.Join(output, "")))
			})
		})

		It("does not format strings when provided no args", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.Say("Hello %s World!") // whoops
			})

			Expect(strings.Join(output, "")).To(Equal("Hello %s World!"))
		})
	})

	Describe("Confirming user input", func() {
		It("treats 'y' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("y\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})

		It("treats other input as a negative confirmation", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Hello World?"}))
			})
		})
	})

	Describe("Confirming deletion", func() {
		It("treats 'y' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("y\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"modelType modelName"}))
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeTrue())
				})

				Expect(out).To(ContainSubstrings([]string{"modelType modelName"}))
			})
		})

		It("treats other input as a negative confirmation and warns the user", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Delete cancelled"}))
			})
		})
	})

	Describe("Confirming deletion with associations", func() {
		It("warns the user that associated objects will also be deleted", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDeleteWithAssociations("modelType", "modelName")).To(BeFalse())
				})

				Expect(out).To(ContainSubstrings([]string{"Delete cancelled"}))
			})
		})
	})

	Context("when user is not logged in", func() {
		var config configuration.Reader

		BeforeEach(func() {
			config = testconfig.NewRepository()
		})

		It("prompts the user to login", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI((os.Stdin))
				ui.ShowConfiguration(config)
			})

			Expect(output).ToNot(ContainSubstrings([]string{"API endpoint:"}))
			Expect(output).To(ContainSubstrings([]string{"Not logged in", "Use", "log in"}))
		})
	})

	Context("when an api endpoint is set and the user logged in", func() {
		var config configuration.ReadWriter

		BeforeEach(func() {
			accessToken := configuration.TokenInfo{
				UserGuid: "my-user-guid",
				Username: "my-user",
				Email:    "my-user-email",
			}
			config = testconfig.NewRepositoryWithAccessToken(accessToken)
			config.SetApiEndpoint("https://test.example.org")
			config.SetApiVersion("☃☃☃")
		})

		Describe("tells the user what is set in the config", func() {
			var output []string

			JustBeforeEach(func() {
				output = io_helpers.CaptureOutput(func() {
					ui := NewUI(os.Stdin)
					ui.ShowConfiguration(config)
				})
			})

			It("tells the user which api endpoint is set", func() {
				Expect(output).To(ContainSubstrings([]string{"API endpoint:", "https://test.example.org"}))
			})

			It("tells the user the api version", func() {
				Expect(output).To(ContainSubstrings([]string{"API version:", "☃☃☃"}))
			})

			It("tells the user which user is logged in", func() {
				Expect(output).To(ContainSubstrings([]string{"User:", "my-user-email"}))
			})

			Context("when an org is targeted", func() {
				BeforeEach(func() {
					config.SetOrganizationFields(models.OrganizationFields{
						Name: "org-name",
						Guid: "org-guid",
					})
				})

				It("tells the user which org is targeted", func() {
					Expect(output).To(ContainSubstrings([]string{"Org:", "org-name"}))
				})
			})

			Context("when a space is targeted", func() {
				BeforeEach(func() {
					config.SetSpaceFields(models.SpaceFields{
						Name: "my-space",
						Guid: "space-guid",
					})
				})

				It("tells the user which space is targeted", func() {
					Expect(output).To(ContainSubstrings([]string{"Space:", "my-space"}))
				})
			})
		})

		It("prompts the user to target an org and space when no org or space is targeted", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"}))
		})

		It("prompts the user to target an org when no org is targeted", func() {
			sf := models.SpaceFields{}
			sf.Guid = "guid"
			sf.Name = "name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "org", "targeted", "-o ORG"}))
		})

		It("prompts the user to target a space when no space is targeted", func() {
			of := models.OrganizationFields{}
			of.Guid = "of-guid"
			of.Name = "of-name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			Expect(output).To(ContainSubstrings([]string{"No", "space", "targeted", "-s SPACE"}))
		})
	})

	Describe("failing", func() {
		It("panics with a specific string", func() {
			io_helpers.CaptureOutput(func() {
				testassert.AssertPanic(FailedWasCalled, func() {
					NewUI(os.Stdin).Failed("uh oh")
				})
			})
		})
	})
})
