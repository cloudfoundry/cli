package errors_test

import (
	. "github.com/cloudfoundry/cli/cf/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Specific Errors", func() {
	Describe("AccessDeniedError", func() {
		It("creates an access denied error", func() {
			err := NewAccessDeniedError()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("403"))
			Expect(err.Error()).To(ContainSubstring("Access is denied"))
		})
	})

	Describe("AsyncTimeoutError", func() {
		It("creates an async timeout error with URL", func() {
			err := NewAsyncTimeoutError("http://example.com/job/123")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timed out"))
			Expect(err.Error()).To(ContainSubstring("http://example.com/job/123"))
		})

		It("handles empty URL", func() {
			err := NewAsyncTimeoutError("")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("HttpError", func() {
		It("creates a base HTTP error", func() {
			err := NewHttpError(500, "SERVER_ERROR", "Internal server error")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("500"))
			Expect(err.Error()).To(ContainSubstring("SERVER_ERROR"))
			Expect(err.Error()).To(ContainSubstring("Internal server error"))
		})

		It("creates HttpNotFoundError for 404 status", func() {
			err := NewHttpError(404, "NOT_FOUND", "Resource not found")
			Expect(err).To(HaveOccurred())
			_, ok := err.(*HttpNotFoundError)
			Expect(ok).To(BeTrue())
		})

		It("implements HttpError interface", func() {
			err := NewHttpError(403, "FORBIDDEN", "Access forbidden")
			httpErr, ok := err.(HttpError)
			Expect(ok).To(BeTrue())
			Expect(httpErr.StatusCode()).To(Equal(403))
			Expect(httpErr.ErrorCode()).To(Equal("FORBIDDEN"))
		})

		It("returns correct status code", func() {
			err := NewHttpError(503, "SERVICE_UNAVAILABLE", "Service unavailable")
			httpErr := err.(HttpError)
			Expect(httpErr.StatusCode()).To(Equal(503))
		})

		It("returns correct error code", func() {
			err := NewHttpError(400, "BAD_REQUEST", "Bad request")
			httpErr := err.(HttpError)
			Expect(httpErr.ErrorCode()).To(Equal("BAD_REQUEST"))
		})
	})

	Describe("InvalidSSLCert", func() {
		It("creates an invalid SSL cert error with URL and reason", func() {
			err := NewInvalidSSLCert("https://example.com", "self-signed certificate")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("https://example.com"))
			Expect(err.Error()).To(ContainSubstring("self-signed certificate"))
			Expect(err.Error()).To(ContainSubstring("invalid SSL certificate"))
		})

		It("handles empty reason", func() {
			err := NewInvalidSSLCert("https://example.com", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("https://example.com"))
			Expect(err.Error()).ToNot(ContainSubstring(" - "))
		})

		It("includes both URL and reason when provided", func() {
			err := NewInvalidSSLCert("https://api.example.com", "certificate expired")
			Expect(err.URL).To(Equal("https://api.example.com"))
			Expect(err.Reason).To(Equal("certificate expired"))
		})
	})

	Describe("InvalidTokenError", func() {
		It("creates an invalid token error", func() {
			err := NewInvalidTokenError("token expired")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid auth token"))
			Expect(err.Error()).To(ContainSubstring("token expired"))
		})

		It("handles empty description", func() {
			err := NewInvalidTokenError("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid auth token"))
		})
	})

	Describe("ModelAlreadyExistsError", func() {
		It("creates a model already exists error", func() {
			err := NewModelAlreadyExistsError("space", "my-space")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("space"))
			Expect(err.Error()).To(ContainSubstring("my-space"))
			Expect(err.Error()).To(ContainSubstring("already exists"))
		})

		It("stores model type and name", func() {
			err := NewModelAlreadyExistsError("organization", "my-org")
			Expect(err.ModelType).To(Equal("organization"))
			Expect(err.ModelName).To(Equal("my-org"))
		})

		It("handles different model types", func() {
			err := NewModelAlreadyExistsError("route", "example.com")
			Expect(err.Error()).To(ContainSubstring("route"))
			Expect(err.Error()).To(ContainSubstring("example.com"))
		})
	})

	Describe("ModelNotFoundError", func() {
		It("creates a model not found error", func() {
			err := NewModelNotFoundError("app", "my-app")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("app"))
			Expect(err.Error()).To(ContainSubstring("my-app"))
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("formats the message correctly", func() {
			err := NewModelNotFoundError("service", "my-service")
			modelErr, ok := err.(*ModelNotFoundError)
			Expect(ok).To(BeTrue())
			Expect(modelErr.ModelType).To(Equal("service"))
			Expect(modelErr.ModelName).To(Equal("my-service"))
		})
	})

	Describe("NotAuthorizedError", func() {
		It("creates a not authorized error", func() {
			err := NewNotAuthorizedError()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403"))
			Expect(err.Error()).To(ContainSubstring("10003"))
			Expect(err.Error()).To(ContainSubstring("not authorized"))
		})
	})

	Describe("EmptyDirError", func() {
		It("creates an empty directory error", func() {
			err := NewEmptyDirError("/path/to/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("/path/to/dir"))
			Expect(err.Error()).To(ContainSubstring("is empty"))
		})

		It("handles relative paths", func() {
			err := NewEmptyDirError("./my-dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("./my-dir"))
		})
	})

	Describe("ServiceAssociationError", func() {
		It("creates a service association error", func() {
			err := NewServiceAssociationError()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Cannot delete service instance"))
			Expect(err.Error()).To(ContainSubstring("service keys and bindings"))
		})
	})

	Describe("UnbindableServiceError", func() {
		It("creates an unbindable service error", func() {
			err := NewUnbindableServiceError()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("doesn't support creation of keys"))
		})
	})
})
