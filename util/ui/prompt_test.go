package ui_test

import (
	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/util/ui/uifakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/vito/go-interact/interact"
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
		Describe("display text", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("displays the passed in string", func() {
				_, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Say(`some-prompt \[yN\]:`))
			})
		})

		When("the user chooses yes", func() {
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

		When("the user chooses no", func() {
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

		When("the user chooses the default", func() {
			BeforeEach(func() {
				_, err := inBuffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			When("the default is true", func() {
				It("returns true", func() {
					response, err := ui.DisplayBoolPrompt(true, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeTrue())
				})
			})

			When("the default is false", func() {
				It("returns false", func() {
					response, err := ui.DisplayBoolPrompt(false, "some-prompt", nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(response).To(BeFalse())
				})
			})
		})

		When("the interact library returns an error", func() {
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
			_, err := inBuffer.Write([]byte("some-input\n"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("displays the passed in string", func() {
			_, err := ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})
			Expect(err).NotTo(HaveOccurred())
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

		When("the locale is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
				ui.OutForInteration = out

				inBuffer = NewBuffer()
				ui.In = inBuffer

				_, err = inBuffer.Write([]byte("ffffff\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("translates and displays the prompt", func() {
				_, err := ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Say("L'application some-app n'existe pas.\n"))
			})
		})
	})

	Describe("DisplayTextPrompt", func() {
		BeforeEach(func() {
			_, err := inBuffer.Write([]byte("some-input\n"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("displays the passed in string and returns the user input string", func() {
			userInput, err := ui.DisplayTextPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
				"AppName": "some-app",
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Say("App some-app does not exist."))
			Expect(userInput).To(Equal("some-input"))
			Expect(out).To(Say("some-input"))
		})

		When("the local is not set to English", func() {
			BeforeEach(func() {
				fakeConfig.LocaleReturns("fr-FR")

				var err error
				ui, err = NewUI(fakeConfig)
				Expect(err).NotTo(HaveOccurred())

				ui.Out = out
				ui.OutForInteration = out

				inBuffer = NewBuffer()
				ui.In = inBuffer

				_, err = inBuffer.Write([]byte("ffffff\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("translates and displays the prompt", func() {
				_, err := ui.DisplayTextPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Say("L'application some-app n'existe pas.\n"))
			})
		})
	})

	Describe("interrupt handling", func() {
		When("the prompt is canceled by a keyboard interrupt", func() {
			var (
				fakeResolver   *uifakes.FakeResolver
				fakeExiter     *uifakes.FakeExiter
				fakeInteractor *uifakes.FakeInteractor
			)

			BeforeEach(func() {
				fakeResolver = new(uifakes.FakeResolver)
				fakeResolver.ResolveReturns(interact.ErrKeyboardInterrupt)
				fakeExiter = new(uifakes.FakeExiter)
				fakeInteractor = new(uifakes.FakeInteractor)
				fakeInteractor.NewInteractionReturns(fakeResolver)
				ui.Interactor = fakeInteractor
				ui.Exiter = fakeExiter
			})

			It("exits immediately from password prompt", func() {
				_, _ = ui.DisplayPasswordPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(fakeExiter.ExitCallCount()).To(Equal(1))
				Expect(fakeExiter.ExitArgsForCall(0)).To(Equal(130))
			})

			It("exits immediately from text prompt", func() {
				_, _ = ui.DisplayTextPrompt("App {{.AppName}} does not exist.", map[string]interface{}{
					"AppName": "some-app",
				})
				Expect(fakeExiter.ExitCallCount()).To(Equal(1))
				Expect(fakeExiter.ExitArgsForCall(0)).To(Equal(130))
			})

			It("exits immediately from bool prompt", func() {
				_, _ = ui.DisplayBoolPrompt(false, "some-prompt", nil)
				Expect(fakeExiter.ExitCallCount()).To(Equal(1))
				Expect(fakeExiter.ExitArgsForCall(0)).To(Equal(130))
			})
		})

	})
})
