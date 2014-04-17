package terminal_test

import (
	"cf/configuration"
	"cf/io_helpers"
	"cf/models"
	. "cf/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"strings"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
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

				testassert.SliceContains(out, testassert.Lines{{"Hello World?"}})
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeTrue())
				})

				testassert.SliceContains(out, testassert.Lines{{"Hello World?"}})
			})
		})

		It("treats other input as a negative confirmation", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.Confirm("Hello %s", "World?")).To(BeFalse())
				})

				testassert.SliceContains(out, testassert.Lines{{"Hello World?"}})
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

				testassert.SliceContains(out, testassert.Lines{{"modelType modelName"}})
			})
		})

		It("treats 'yes' as an affirmative confirmation", func() {
			io_helpers.SimulateStdin("yes\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeTrue())
				})

				testassert.SliceContains(out, testassert.Lines{{"modelType modelName"}})
			})
		})

		It("treats other input as a negative confirmation and warns the user", func() {
			io_helpers.SimulateStdin("wat\n", func(reader io.Reader) {
				out := io_helpers.CaptureOutput(func() {
					ui := NewUI(reader)
					Expect(ui.ConfirmDelete("modelType", "modelName")).To(BeFalse())
				})

				testassert.SliceContains(out, testassert.Lines{{"Delete cancelled"}})
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

				testassert.SliceContains(out, testassert.Lines{{"Delete cancelled"}})
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

			testassert.SliceDoesNotContain(output, testassert.Lines{
				{"API endpoint:"},
			})

			testassert.SliceContains(output, testassert.Lines{
				{"Not logged in", "Use", "log in"},
			})
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
				testassert.SliceContains(output, testassert.Lines{
					{"API endpoint:", "https://test.example.org"},
				})
			})

			It("tells the user the api version", func() {
				testassert.SliceContains(output, testassert.Lines{
					{"API version:", "☃☃☃"},
				})
			})

			It("tells the user which user is logged in", func() {
				testassert.SliceContains(output, testassert.Lines{
					{"User:", "my-user-email"},
				})
			})

			Context("when an org is targeted", func() {
				BeforeEach(func() {
					config.SetOrganizationFields(models.OrganizationFields{
						Name: "org-name",
						Guid: "org-guid",
					})
				})

				It("tells the user which org is targeted", func() {
					testassert.SliceContains(output, testassert.Lines{
						{"Org:", "org-name"},
					})
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
					testassert.SliceContains(output, testassert.Lines{
						{"Space:", "my-space"},
					})
				})
			})
		})

		It("prompts the user to target an org and space when no org or space is targeted", func() {
			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "org", "space", "targeted", "-o ORG", "-s SPACE"},
			})
		})

		It("prompts the user to target an org when no org is targeted", func() {
			sf := models.SpaceFields{}
			sf.Guid = "guid"
			sf.Name = "name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "org", "targeted", "-o ORG"},
			})
		})

		It("prompts the user to target a space when no space is targeted", func() {
			of := models.OrganizationFields{}
			of.Guid = "of-guid"
			of.Name = "of-name"

			output := io_helpers.CaptureOutput(func() {
				ui := NewUI(os.Stdin)
				ui.ShowConfiguration(config)
			})

			testassert.SliceContains(output, testassert.Lines{
				{"No", "space", "targeted", "-s SPACE"},
			})
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
