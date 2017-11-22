package rata_test

import (
	. "github.com/tedsuo/rata"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requests", func() {
	var (
		host             string
		requestGenerator *RequestGenerator
	)
	const (
		PathWithSlash    = "WithSlash"
		PathWithoutSlash = "WithoutSlash"
		PathWithParams   = "WithParams"
	)

	JustBeforeEach(func() {
		routes := Routes{
			{Name: PathWithSlash, Method: "GET", Path: "/some-route"},
			{Name: PathWithoutSlash, Method: "GET", Path: "some-route"},
			{Name: PathWithParams, Method: "GET", Path: "/foo/:bar"},
		}
		requestGenerator = NewRequestGenerator(
			host,
			routes,
		)
	})

	Describe("CreateRequest", func() {
		Context("when the host does not have a trailing slash", func() {
			BeforeEach(func() {
				host = "http://example.com"
			})

			Context("when the path starts with a slash", func() {
				It("generates a URL with one slash between the host and the path", func() {
					request, err := requestGenerator.CreateRequest(PathWithSlash, Params{}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(request.URL.String()).To(Equal("http://example.com/some-route"))
				})
			})

			Context("when the path does not start with a slash", func() {
				It("generates a URL with one slash between the host and the path", func() {
					request, err := requestGenerator.CreateRequest(PathWithoutSlash, Params{}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(request.URL.String()).To(Equal("http://example.com/some-route"))
				})
			})
		})

		Context("when host has a trailing slash", func() {
			BeforeEach(func() {
				host = "example.com/"
			})

			Context("when the path starts with a slash", func() {
				It("generates a URL with one slash between the host and the path", func() {
					request, err := requestGenerator.CreateRequest(PathWithSlash, Params{}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(request.URL.String()).To(Equal("example.com/some-route"))
				})
			})

			Context("when the path does not start with a slash", func() {
				It("generates a URL with one slash between the host and the path", func() {
					request, err := requestGenerator.CreateRequest(PathWithoutSlash, Params{}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(request.URL.String()).To(Equal("example.com/some-route"))
				})
			})
		})

		Context("when using params that can be interpreted as path segments", func() {
			Context("without a host", func() {
				BeforeEach(func() {
					host = ""
				})

				It("generates a URL with the slash encoded as %2F", func() {
					request, err := requestGenerator.CreateRequest(PathWithParams, Params{"bar": "something/with/slashes"}, nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(request.URL.String()).To(Equal("/foo/something%2Fwith%2Fslashes"))
				})
			})

			Context("with a host", func() {
				BeforeEach(func() {
					host = "http://example.com/test"
				})

				Context("when the param has a slash", func() {
					It("generates a URL with the slash encoded as %2F", func() {
						request, err := requestGenerator.CreateRequest(PathWithParams, Params{"bar": "something/with/slashes"}, nil)
						Expect(err).NotTo(HaveOccurred())

						Expect(request.URL.String()).To(Equal("http://example.com/test/foo/something%2Fwith%2Fslashes"))
					})
				})

				Context("when the param has a question mark", func() {
					It("generates a URL with the question mark encoded as %3F", func() {
						request, err := requestGenerator.CreateRequest(PathWithParams, Params{"bar": "something?with?question?marks"}, nil)
						Expect(err).NotTo(HaveOccurred())

						Expect(request.URL.String()).To(Equal("http://example.com/test/foo/something%3Fwith%3Fquestion%3Fmarks"))
					})
				})

				Context("when the param has a pound sign", func() {
					It("generates a URL with the pound sign encoded as %23", func() {
						request, err := requestGenerator.CreateRequest(PathWithParams, Params{"bar": "something#with#pound#signs"}, nil)
						Expect(err).NotTo(HaveOccurred())

						Expect(request.URL.String()).To(Equal("http://example.com/test/foo/something%23with%23pound%23signs"))
					})
				})
			})
		})
	})
})
