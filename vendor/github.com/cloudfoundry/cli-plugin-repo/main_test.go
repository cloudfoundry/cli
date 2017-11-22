package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"io/ioutil"
	"time"

	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Integration", func() {
	var (
		session *gexec.Session
		err     error
		port    string
	)

	Context("--force-ssl not set", func() {
		BeforeEach(func() {
			port = strconv.Itoa(8080 + GinkgoParallelNode())
			session, err = gexec.Start(
				exec.Command(buildPath, "-p", port, "-f", "fixtures/repo-index.yml"),
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)
		})

		AfterEach(func() {
			session.Kill()
		})

		Describe("/", func() {
			It("returns HTML we expect", func() {
				client := http.DefaultClient
				response, err := client.Get("http://127.0.0.1:" + port)
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(BeSuccessful())

				b, err := ioutil.ReadFile("ui/index.html")
				Expect(err).NotTo(HaveOccurred())

				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(contents)).To(Equal(string(b)))
			})
		})

		Describe("/list", func() {
			It("returns json that looks like we expect it", func() {
				client := http.DefaultClient
				response, err := client.Get("http://127.0.0.1:" + port + "/list")
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(BeSuccessful())

				b, err := ioutil.ReadFile("fixtures/repo-index-response.json")
				Expect(err).NotTo(HaveOccurred())

				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(contents)).To(Equal(string(b)))
			})
		})

		Describe("/ui", func() {
			It("redirects to index", func() {
				client := http.DefaultClient
				response, err := client.Get("http://127.0.0.1:" + port + "/ui")
				Expect(err).NotTo(HaveOccurred())
				Expect(response).To(BeSuccessful())

				Expect(response.Request.URL.Path).To(Equal("/"))
			})
		})
	})

	Context("--force-ssl is set", func() {
		BeforeEach(func() {
			port = strconv.Itoa(8080 + GinkgoParallelNode())
			session, err = gexec.Start(
				exec.Command(buildPath, "-p", port, "-f", "fixtures/repo-index.yml", "--force-ssl"),
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)
		})

		AfterEach(func() {
			session.Kill()
		})

		Context("when 'x-forwarded-proto' is set to 'http'", func() {
			Describe("/", func() {
				It("redirects to the https url", func() {
					transport := http.Transport{}
					request, err := http.NewRequest("GET", "http://127.0.0.1:"+port, nil)
					Expect(err).NotTo(HaveOccurred())
					request.Header.Set("x-forwarded-proto", "http")

					response, err := transport.RoundTrip(request)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(BeSuccessful())

					redirectLocation, err := response.Location()
					Expect(err).NotTo(HaveOccurred())
					Expect(redirectLocation).To(MatchRegexp("^https:"))
				})
			})

			Describe("/list", func() {
				It("redirects to the https url", func() {
					transport := http.Transport{}
					request, err := http.NewRequest("GET", "http://127.0.0.1:"+port+"/list", nil)
					Expect(err).NotTo(HaveOccurred())
					request.Header.Set("x-forwarded-proto", "http")

					response, err := transport.RoundTrip(request)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(BeSuccessful())

					redirectLocation, err := response.Location()
					Expect(err).NotTo(HaveOccurred())
					Expect(redirectLocation).To(MatchRegexp("^https:"))
				})
			})

			Describe("/ui", func() {
				It("redirects to the https url", func() {
					transport := http.Transport{}
					request, err := http.NewRequest("GET", "http://127.0.0.1:"+port+"/ui", nil)
					Expect(err).NotTo(HaveOccurred())
					request.Header.Set("x-forwarded-proto", "http")

					response, err := transport.RoundTrip(request)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(BeSuccessful())

					redirectLocation, err := response.Location()
					Expect(err).NotTo(HaveOccurred())
					Expect(redirectLocation).To(MatchRegexp("^https:"))
				})
			})

			Describe("https request", func() {
				It("does not do a redirect", func() {
					transport := http.Transport{}
					request, err := http.NewRequest("GET", "http://127.0.0.1:"+port, nil)
					Expect(err).NotTo(HaveOccurred())
					request.Header.Set("x-forwarded-proto", "https")

					response, err := transport.RoundTrip(request)
					Expect(err).NotTo(HaveOccurred())
					Expect(response).To(BeSuccessful())

					_, err = response.Location()
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func BeSuccessful() types.GomegaMatcher {
	return &SuccessfulHTTPResponseMatcher{}
}

type SuccessfulHTTPResponseMatcher struct{}

func (matcher *SuccessfulHTTPResponseMatcher) Match(actual interface{}) (success bool, err error) {
	response, ok := actual.(*http.Response)
	if !ok {
		return false, fmt.Errorf("SuccessfulHTTPResponseMatcher matcher expects an http.Response")
	}

	return (response.StatusCode >= 200) && (response.StatusCode < 400), nil
}

func (matcher *SuccessfulHTTPResponseMatcher) FailureMessage(actual interface{}) (message string) {
	response := actual.(*http.Response)

	return fmt.Sprintf("Expected Status Code\n\t%d\nto be successful (2XX or 3XX)", response.StatusCode)
}

func (matcher *SuccessfulHTTPResponseMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	response := actual.(*http.Response)

	return fmt.Sprintf("Expected Status Code\n\t%d\nto not be successful (1XX, 4XX, 5XX)", response.StatusCode)
}
