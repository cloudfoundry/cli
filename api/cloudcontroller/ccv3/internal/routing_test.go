package internal_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routing", func() {
	Describe("Route", func() {
		var route Route

		Describe("CreatePath", func() {
			BeforeEach(func() {
				route = Route{
					Name:   "whatevz",
					Method: "GET",
					Path:   "/a/path/:param/with/:many_things/:many/in/:it",
				}
			})

			It("should return a url with all :entries populated by the passed in hash", func() {
				Expect(route.CreatePath(Params{
					"param":       "1",
					"many_things": "2",
					"many":        "a space",
					"it":          "4",
				})).Should(Equal(`/a/path/1/with/2/a%20space/in/4`))
			})

			Context("when the hash is missing params", func() {
				It("should error", func() {
					_, err := route.CreatePath(Params{
						"param": "1",
						"many":  "2",
						"it":    "4",
					})
					Expect(err).Should(HaveOccurred())
				})
			})

			Context("when the hash has extra params", func() {
				It("should totally not care", func() {
					Expect(route.CreatePath(Params{
						"param":       "1",
						"many_things": "2",
						"many":        "a space",
						"it":          "4",
						"donut":       "bacon",
					})).Should(Equal(`/a/path/1/with/2/a%20space/in/4`))
				})
			})

			Context("with a trailing slash", func() {
				It("should work", func() {
					route = Route{
						Name:   "whatevz",
						Method: "GET",
						Path:   "/a/path/:param/",
					}
					Expect(route.CreatePath(Params{
						"param": "1",
					})).Should(Equal(`/a/path/1/`))
				})
			})
		})
	})

	Describe("Router", func() {
		var (
			router    *Router
			routes    []Route
			resources map[string]string
		)

		JustBeforeEach(func() {
			router = NewRouter(routes, resources)
		})

		Describe("CreateRequest", func() {
			Context("when the route exists", func() {
				var badRouteName, routeName string
				BeforeEach(func() {
					routeName = "banana"
					badRouteName = "orange"

					routes = []Route{
						{Name: routeName, Resource: "exists", Path: "/very/good/:name", Method: http.MethodGet},
						{Name: badRouteName, Resource: "fake-resource", Path: "/very/bad", Method: http.MethodGet},
					}
				})

				Context("when the resource exists exists", func() {
					BeforeEach(func() {
						resources = map[string]string{
							"exists": "https://foo.bar.baz/this/is",
						}
					})

					It("returns a request", func() {
						request, err := router.CreateRequest(routeName, Params{"name": "Henry the 8th"}, nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(request.URL.String()).To(Equal("https://foo.bar.baz/this/is/very/good/Henry%2520the%25208th"))
					})
				})

				Context("when the resource exists exists", func() {
					It("returns an error", func() {
						_, err := router.CreateRequest(badRouteName, nil, nil)
						Expect(err).To(MatchError("No resource exists with the name fake-resource"))
					})
				})
			})

			Context("when the route does not exists exist", func() {
				It("returns an error", func() {
					_, err := router.CreateRequest("fake-route", nil, nil)
					Expect(err).To(MatchError("No route exists with the name fake-route"))
				})
			})
		})
	})
})
