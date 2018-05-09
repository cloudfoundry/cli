package ui_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Prompts", func() {
	var (
		ui         *UI
		fakeConfig *uifakes.FakeConfig
		out        *Buffer
		errBuff    *Buffer
		inBuffer   *Buffer
	)

	BeforeEach(func() {
		fakeConfig = new(uifakes.FakeConfig)
		fakeConfig.ColorEnabledReturns(configv3.ColorEnabled)

		var err error
		ui, err = NewUI(fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		out = NewBuffer()
		ui.Out = out
		ui.OutForInteration = out
		errBuff = NewBuffer()
		ui.Err = errBuff

		inBuffer = NewBuffer()
		ui.In = inBuffer
	})

	Describe("DisplayBoolPrompt", func() {
		It("displays the passed in string", func() {
			_, _ = ui.DisplayBoolPrompt(false, "some-prompt", nil)
			Expect(out).To(Say("some-prompt \\[yN\\]:"))
		})

		Context("when the user chooses yes", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns true", func() {
				response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeTrue())
			})
		})

		Context("when the user chooses no", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false", func() {
				response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(response).To(BeFalse())
			})
		})

		Context("when the user chooses the default", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the default is true", func() {
				It("returns true", func() {
					response, err := ui.DisplayBoolPrompt(true, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeTrue())
				})
			})

			Context("when the default is false", func() {
				It("returns false", func() {
					response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeFalse())
				})
			})
		})

		Context("when the interact library returns an error", func() {
			It("returns the error", func() {
				_, err := inBuffer.Write([]byte("invalid\n"))
				Expect(err).ToNot(HaveOccurred())
				_, err = ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DisplayPasswordPrompt", func() {
		BeforeEach(func() {
			inBuffer.Write([]byte("some-input\n"))
		})

		It("displays the passed in string", func() {
			_, _ = ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})
			Expect(out).To(Say("App some-app does not exist."))
		})

		It("returns the user input", func() {
			userInput, err := ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(userInput).To(Equal("some-input"))
			Expect(out).ToNot(Say("some-input"))
		})

		Context("when the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
				ui.OutForInteration = out
			})

			It("translates and displays the prompt", func() {
				_, _ = ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(out).To(Say("L'application some-app n'existe pas.\n"))
			})
		})
	})
})
