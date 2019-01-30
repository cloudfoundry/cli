package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/translatableerror/translatableerrorfakes"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TranslateSpy struct {
	wasTranslateCalled bool
	calls              []struct {
		text string
		keys []interface{}
	}
}

func (spy *TranslateSpy) translate(s string, v ...interface{}) string {
	spy.wasTranslateCalled = true
	spy.calls = append(spy.calls, struct {
		text string
		keys []interface{}
	}{text: s, keys: v})

	return fmt.Sprintf("Translate Called: %s -- %v", s, v)
}

func (spy TranslateSpy) translateArgsForCall(i int) (string, []interface{}) {
	return spy.calls[i].text, spy.calls[i].keys
}

var _ = Describe("TipDecoratorError", func() {

	Describe("Translate()", func() {
		var (
			tip    TipDecoratorError
			output string

			spy TranslateSpy
		)

		BeforeEach(func() {
			tip = TipDecoratorError{
				Tip: "I am a {{.Foo}}",
				TipKeys: map[string]interface{}{
					"Foo": "tip",
				},
			}
			spy = TranslateSpy{}
		})

		JustBeforeEach(func() {
			output = tip.Translate(spy.translate)
		})

		When("the base error is translatable error", func() {
			var fakeErr *translatableerrorfakes.FakeTranslatableError

			BeforeEach(func() {
				fakeErr = new(translatableerrorfakes.FakeTranslatableError)
				fakeErr.TranslateReturns("some translated error")

				tip.BaseError = fakeErr
			})

			It("translates the base error", func() {
				Expect(fakeErr.TranslateCallCount()).To(Equal(1))
			})

			It("has output", func() {
				Expect(output).To(Equal("Translate Called: {{.BaseError}}\n\nTIP: {{.Tip}} -- [map[BaseError:some translated error Tip:Translate Called: I am a {{.Foo}} -- [map[Foo:tip]]]]"))
			})
		})

		When("the base error is a generic error", func() {
			var genericError error

			BeforeEach(func() {
				genericError = errors.New("I am an error")
				tip.BaseError = genericError
			})

			It("translates the tip", func() {
				Expect(spy.wasTranslateCalled).To(BeTrue())
				actualFormatString, actualFormatKeys := spy.translateArgsForCall(0)

				Expect(actualFormatString).To(Equal("I am a {{.Foo}}"))
				Expect(actualFormatKeys).To(ConsistOf(map[string]interface{}{
					"Foo": "tip",
				}))
			})

			It("has output", func() {
				Expect(output).To(Equal("Translate Called: {{.BaseError}}\n\nTIP: {{.Tip}} -- [map[BaseError:I am an error Tip:Translate Called: I am a {{.Foo}} -- [map[Foo:tip]]]]"))
			})
		})
	})
})
