package rata_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/rata"
)

var _ = Describe("Router", func() {
	Describe("Route", func() {
		var route rata.Route

		Describe("CreatePath", func() {
			BeforeEach(func() {
				route = rata.Route{
					Name:   "whatevz",
					Method: "GET",
					Path:   "/a/path/:param/with/:many_things/:many/in/:it",
				}
			})

			It("should return a url with all :entries populated by the passed in hash", func() {
				Ω(route.CreatePath(rata.Params{
					"param":       "1",
					"many_things": "2",
					"many":        "a space",
					"it":          "with/slash",
				})).Should(Equal(`/a/path/1/with/2/a%20space/in/with%2Fslash`))
			})

			Context("when the hash is missing params", func() {
				It("should error", func() {
					_, err := route.CreatePath(rata.Params{
						"param": "1",
						"many":  "2",
						"it":    "4",
					})
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("when the hash has extra params", func() {
				It("should totally not care", func() {
					Ω(route.CreatePath(rata.Params{
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
					route = rata.Route{
						Name:   "whatevz",
						Method: "GET",
						Path:   "/a/path/:param/",
					}
					Ω(route.CreatePath(rata.Params{
						"param": "1",
					})).Should(Equal(`/a/path/1/`))
				})
			})
		})
	})

	Describe("Routes", func() {
		var routes rata.Routes

		Describe("FindRouteByName", func() {
			BeforeEach(func() {
				routes = rata.Routes{
					{Path: "/something", Method: "GET", Name: "getter"},
					{Path: "/something", Method: "POST", Name: "poster"},
					{Path: "/something", Method: "PuT", Name: "putter"},
					{Path: "/something", Method: "DELETE", Name: "deleter"},
				}
			})

			Context("when the route is present", func() {
				It("returns the route with the matching handler name", func() {
					route, ok := routes.FindRouteByName("getter")
					Ω(ok).Should(BeTrue())
					Ω(route.Method).Should(Equal("GET"))
				})
			})

			Context("when the route is not present", func() {
				It("returns falseness", func() {
					route, ok := routes.FindRouteByName("orangutanger")
					Ω(ok).Should(BeFalse())
					Ω(route).Should(BeZero())
				})
			})
		})

		Describe("PathForName", func() {
			BeforeEach(func() {
				routes = rata.Routes{
					{
						Name:   "whatevz",
						Method: "GET",
						Path:   "/a/path/:param/with/:many_things/:many/in/:it",
					},
				}
			})

			Context("when the route is present", func() {
				It("returns the route with the matching handler name", func() {
					path, err := routes.CreatePathForRoute("whatevz", rata.Params{
						"param":       "1",
						"many_things": "2",
						"many":        "a space",
						"it":          "4",
					})
					Expect(err).NotTo(HaveOccurred())
					Ω(path).Should(Equal(`/a/path/1/with/2/a%20space/in/4`))
				})

				Context("when the route is not present", func() {
					It("returns an error", func() {
						_, err := routes.CreatePathForRoute("foo", rata.Params{
							"param":       "1",
							"many_things": "2",
							"many":        "a space",
							"it":          "4",
						})
						Expect(err).To(HaveOccurred())
					})
				})

				Context("when the hash is missing params", func() {
					It("should error", func() {
						_, err := routes.CreatePathForRoute("whatevz", rata.Params{
							"param": "1",
							"many":  "2",
							"it":    "4",
						})
						Ω(err).Should(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Router", func() {
		var r http.Handler
		var err error
		var routes = rata.Routes{
			{Path: "/something", Method: "GET", Name: "getter"},
			{Path: "/something", Method: "POST", Name: "poster"},
			{Path: "/something", Method: "PuT", Name: "putter"},
			{Path: "/something", Method: "DELETE", Name: "deleter"},
			{Path: "/something", Method: "OPTIONS", Name: "optioner"},
			{Path: "/something", Method: "patch", Name: "patcher"},
			{Path: "/something", Method: "HEAD", Name: "header"},
			{Path: "/something", Method: "CONNEcT", Name: "connector"},
			{Path: "/something", Method: "TRACE", Name: "tracer"},
		}

		Context("when all the handlers are present", func() {
			var resp *httptest.ResponseRecorder
			var handlers = rata.Handlers{
				"getter":    ghttp.RespondWith(http.StatusOK, "get response"),
				"poster":    ghttp.RespondWith(http.StatusOK, "post response"),
				"putter":    ghttp.RespondWith(http.StatusOK, "put response"),
				"deleter":   ghttp.RespondWith(http.StatusOK, "delete response"),
				"optioner":  ghttp.RespondWith(http.StatusOK, "options response"),
				"patcher":   ghttp.RespondWith(http.StatusOK, "patch response"),
				"header":    ghttp.RespondWith(http.StatusOK, "head response"),
				"connector": ghttp.RespondWith(http.StatusOK, "connect response"),
				"tracer":    ghttp.RespondWith(http.StatusOK, "trace response"),
			}

			BeforeEach(func() {
				resp = httptest.NewRecorder()
				r, err = rata.NewRouter(routes, handlers)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("makes GET handlers", func() {
				req, err := http.NewRequest("GET", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("get response"))
			})

			It("makes POST handlers", func() {
				req, err := http.NewRequest("POST", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("post response"))
			})

			It("makes PUT handlers", func() {
				req, err := http.NewRequest("PUT", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("put response"))
			})

			It("makes DELETE handlers", func() {
				req, err := http.NewRequest("DELETE", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("delete response"))
			})

			It("makes OPTIONS handlers", func() {
				req, err := http.NewRequest("OPTIONS", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("options response"))
			})

			It("makes PATCH handlers", func() {
				req, err := http.NewRequest("PATCH", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("patch response"))
			})

			It("makes HEAD handlers", func() {
				req, err := http.NewRequest("HEAD", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("head response"))
			})

			It("makes CONNECT handlers", func() {
				req, err := http.NewRequest("CONNECT", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("connect response"))
			})

			It("makes TRACE handlers", func() {
				req, err := http.NewRequest("TRACE", "/something", nil)
				Ω(err).ShouldNot(HaveOccurred())

				r.ServeHTTP(resp, req)
				Ω(resp.Body.String()).Should(Equal("trace response"))
			})
		})

		Context("when a handler is missing", func() {
			var incompleteHandlers = rata.Handlers{
				"getter": ghttp.RespondWith(http.StatusOK, "get response"),
			}
			It("should error", func() {
				r, err = rata.NewRouter(routes, incompleteHandlers)

				Ω(err).Should(HaveOccurred())
			})
		})

		Context("with an invalid method", func() {
			var invalidRoutes = rata.Routes{
				{Path: "/something", Method: "SMELL", Name: "smeller"},
			}

			It("should error", func() {
				handlers := rata.Handlers{
					"smeller": ghttp.RespondWith(http.StatusOK, "smell response"),
				}
				r, err = rata.NewRouter(invalidRoutes, handlers)

				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Describe("parsing params", func() {
		// this is basically done for us by PAT we simply want to verify some assumptions in these tests
		Context("when all the handlers are present", func() {
			var resp *httptest.ResponseRecorder
			var handlers = rata.Handlers{
				"getter": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte(rata.Param(req, "neato")))
				}),
			}

			Context("when a named path param is provided", func() {
				var r http.Handler
				var err error
				var routes = rata.Routes{
					{Path: "/something/:neato", Method: "GET", Name: "getter"},
				}

				BeforeEach(func() {
					resp = httptest.NewRecorder()
					r, err = rata.NewRouter(routes, handlers)
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("the path param is returned", func() {
					req, _ := http.NewRequest("GET", "/something/the-param-value", nil)

					r.ServeHTTP(resp, req)
					Ω(resp.Body.String()).Should(Equal("the-param-value"))
				})
			})

			Context("when a query param is provided that conflicts with a named path param", func() {
				var r http.Handler
				var err error
				var routes = rata.Routes{
					{Path: "/something/:neato", Method: "GET", Name: "getter"},
				}

				BeforeEach(func() {
					resp = httptest.NewRecorder()
					r, err = rata.NewRouter(routes, handlers)
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("the path param takes precedence", func() {
					req, err := http.NewRequest("GET", "/something/the-param-value?:neato=the-query-value", nil)
					Ω(err).ShouldNot(HaveOccurred())

					r.ServeHTTP(resp, req)
					Ω(resp.Body.String()).Should(Equal("the-param-value"))
				})
			})
		})
	})
})
