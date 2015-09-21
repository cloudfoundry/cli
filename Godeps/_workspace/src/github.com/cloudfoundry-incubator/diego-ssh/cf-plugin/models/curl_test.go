package models_test

import (
	"encoding/json"
	"errors"

	. "github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models"
	"github.com/cloudfoundry/cli/plugin/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Curl", func() {
	var fakeCliConnection *fakes.FakeCliConnection

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
	})

	Context("with a valid response", func() {
		type MyStruct struct {
			SomeString string
		}

		BeforeEach(func() {
			input := []string{"{", `"somestring": "foo"`, "}"}
			fakeCliConnection.CliCommandWithoutTerminalOutputReturns(input, nil)
		})

		It("unmarshals a successful response", func() {
			var result MyStruct
			err := Curl(fakeCliConnection, &result, "a", "b", "c")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.SomeString).To(Equal("foo"))

			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputCallCount()).To(Equal(1))
			Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)).To(Equal([]string{"curl", "a", "b", "c"}))
		})

		It("succeeds with no response object", func() {
			err := Curl(fakeCliConnection, nil, "a", "b", "c")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the cli errors", func() {
		BeforeEach(func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputReturns(nil, errors.New("an error"))
		})

		It("returns an error", func() {
			err := Curl(fakeCliConnection, nil, "a", "b", "c")
			Expect(err).To(MatchError("an error"))
		})
	})

	Context("when the response cannot be unmarshalled", func() {
		BeforeEach(func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{"abcd"}, nil)
		})

		It("returns an error", func() {
			err := Curl(fakeCliConnection, nil, "a", "b", "c")
			Expect(err).To(BeAssignableToTypeOf(&json.SyntaxError{}))
		})
	})

	Context("when the response is a CF Error", func() {
		BeforeEach(func() {
			fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{"{", `"code": 1234,`, `"description":"another error"`, "}"}, nil)
		})

		It("returns an error", func() {
			err := Curl(fakeCliConnection, nil, "a", "b", "c")
			Expect(err).To(MatchError("another error"))
		})
	})
})
