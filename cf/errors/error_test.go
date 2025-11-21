package errors_test

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error", func() {
	Describe("New", func() {
		It("creates a new error with the given message", func() {
			err := New("test error message")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("test error message"))
		})

		It("creates different errors for different messages", func() {
			err1 := New("error 1")
			err2 := New("error 2")
			Expect(err1.Error()).ToNot(Equal(err2.Error()))
		})
	})

	Describe("NewWithFmt", func() {
		It("creates a formatted error message", func() {
			err := NewWithFmt("error code: %d, message: %s", 404, "not found")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error code: 404, message: not found"))
		})

		It("handles multiple format arguments", func() {
			err := NewWithFmt("%s %s %d", "first", "second", 3)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("first second 3"))
		})

		It("handles no arguments", func() {
			err := NewWithFmt("simple message")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("simple message"))
		})
	})

	Describe("NewWithError", func() {
		It("wraps an existing error with a message", func() {
			originalErr := errors.New("original error")
			err := NewWithError("wrapped", originalErr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("wrapped: original error"))
		})

		It("combines the message and error correctly", func() {
			originalErr := errors.New("connection failed")
			err := NewWithError("Failed to connect to server", originalErr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to connect to server"))
			Expect(err.Error()).To(ContainSubstring("connection failed"))
		})
	})

	Describe("NewWithSlice", func() {
		It("combines multiple errors into one message", func() {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			err3 := errors.New("error 3")

			err := NewWithSlice([]error{err1, err2, err3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error 1"))
			Expect(err.Error()).To(ContainSubstring("error 2"))
			Expect(err.Error()).To(ContainSubstring("error 3"))
		})

		It("handles empty slice", func() {
			err := NewWithSlice([]error{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(""))
		})

		It("handles single error", func() {
			singleErr := errors.New("single error")
			err := NewWithSlice([]error{singleErr})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("single error"))
		})

		It("formats errors with newlines", func() {
			err1 := errors.New("first")
			err2 := errors.New("second")

			err := NewWithSlice([]error{err1, err2})
			Expect(err.Error()).To(ContainSubstring("\n"))
		})
	})
})
