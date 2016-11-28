package applicationbits_test

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	testapi "code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api/applicationbits"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerApplicationBitsRepository", func() {
	var (
		fixturesDir string
		repo        Repository
		file1       resources.AppFileResource
		file2       resources.AppFileResource
		file3       resources.AppFileResource
		file4       resources.AppFileResource
		testServer  *httptest.Server
		configRepo  coreconfig.ReadWriter
	)

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = filepath.Join(cwd, "../../../fixtures/applications")

		configRepo = testconfig.NewRepositoryWithDefaults()

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		gateway.PollingThrottle = time.Duration(0)

		repo = NewCloudControllerApplicationBitsRepository(configRepo, gateway)

		file1 = resources.AppFileResource{Path: "app.rb", Sha1: "2474735f5163ba7612ef641f438f4b5bee00127b", Size: 51}
		file2 = resources.AppFileResource{Path: "config.ru", Sha1: "f097424ce1fa66c6cb9f5e8a18c317376ec12e05", Size: 70}
		file3 = resources.AppFileResource{Path: "Gemfile", Sha1: "d9c3a51de5c89c11331d3b90b972789f1a14699a", Size: 59, Mode: "0750"}
		file4 = resources.AppFileResource{Path: "Gemfile.lock", Sha1: "345f999aef9070fb9a608e65cf221b7038156b6d", Size: 229, Mode: "0600"}
	})

	setupTestServer := func(reqs ...testnet.TestRequest) {
		testServer, _ = testnet.NewServer(reqs)
		configRepo.SetAPIEndpoint(testServer.URL)
	}

	Describe(".UploadBits", func() {
		var uploadFile *os.File
		var err error

		BeforeEach(func() {
			uploadFile, err = os.Open(filepath.Join(fixturesDir, "ignored_and_resource_matched_example_app.zip"))
			if err != nil {
				log.Fatal(err)
			}
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("uploads zip files", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/apps/my-cool-app-guid/bits",
				Matcher: uploadBodyMatcher(defaultZipCheck),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `
					{
						"metadata":{
							"guid": "my-job-guid",
							"url": "/v2/jobs/my-job-guid"
						}
					}`,
				},
			}),
				createProgressEndpoint("running"),
				createProgressEndpoint("finished"),
			)

			apiErr := repo.UploadBits("my-cool-app-guid", uploadFile, []resources.AppFileResource{file1, file2})
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns a failure when uploading bits fails", func() {
			setupTestServer(testapi.NewCloudControllerTestRequest(testnet.TestRequest{
				Method:  "PUT",
				Path:    "/v2/apps/my-cool-app-guid/bits",
				Matcher: uploadBodyMatcher(defaultZipCheck),
				Response: testnet.TestResponse{
					Status: http.StatusCreated,
					Body: `
					{
						"metadata":{
							"guid": "my-job-guid",
							"url": "/v2/jobs/my-job-guid"
						}
					}`,
				},
			}),
				createProgressEndpoint("running"),
				createProgressEndpoint("failed"),
			)
			apiErr := repo.UploadBits("my-cool-app-guid", uploadFile, []resources.AppFileResource{file1, file2})

			Expect(apiErr).To(HaveOccurred())
		})

		Context("when there are no files to upload", func() {
			It("makes a request without a zipfile", func() {
				setupTestServer(
					testapi.NewCloudControllerTestRequest(testnet.TestRequest{
						Method: "PUT",
						Path:   "/v2/apps/my-cool-app-guid/bits",
						Matcher: func(request *http.Request) {
							err := request.ParseMultipartForm(maxMultipartResponseSizeInBytes)
							Expect(err).NotTo(HaveOccurred())
							defer request.MultipartForm.RemoveAll()

							Expect(len(request.MultipartForm.Value)).To(Equal(1), "Should have 1 value")
							valuePart, ok := request.MultipartForm.Value["resources"]

							Expect(ok).To(BeTrue(), "Resource manifest not present")
							Expect(valuePart).To(Equal([]string{"[]"}))
							Expect(request.MultipartForm.File).To(BeEmpty())
						},
						Response: testnet.TestResponse{
							Status: http.StatusCreated,
							Body: `
					{
						"metadata":{
							"guid": "my-job-guid",
							"url": "/v2/jobs/my-job-guid"
						}
					}`,
						},
					}),
					createProgressEndpoint("running"),
					createProgressEndpoint("finished"),
				)

				apiErr := repo.UploadBits("my-cool-app-guid", nil, []resources.AppFileResource{})
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		It("marshals a nil presentFiles parameter into an empty array", func() {
			setupTestServer(
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method: "PUT",
					Path:   "/v2/apps/my-cool-app-guid/bits",
					Matcher: func(request *http.Request) {
						err := request.ParseMultipartForm(maxMultipartResponseSizeInBytes)
						Expect(err).NotTo(HaveOccurred())
						defer request.MultipartForm.RemoveAll()

						Expect(len(request.MultipartForm.Value)).To(Equal(1), "Should have 1 value")
						valuePart, ok := request.MultipartForm.Value["resources"]

						Expect(ok).To(BeTrue(), "Resource manifest not present")
						Expect(valuePart).To(Equal([]string{"[]"}))
						Expect(request.MultipartForm.File).To(BeEmpty())
					},
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body: `
					{
						"metadata":{
							"guid": "my-job-guid",
							"url": "/v2/jobs/my-job-guid"
						}
					}`,
					},
				}),
				createProgressEndpoint("running"),
				createProgressEndpoint("finished"),
			)

			apiErr := repo.UploadBits("my-cool-app-guid", nil, nil)
			Expect(apiErr).NotTo(HaveOccurred())
		})
	})

	Describe(".GetApplicationFiles", func() {
		It("accepts a slice of files and returns a slice of the files that it already has", func() {
			setupTestServer(matchResourceRequest)
			matchedFiles, err := repo.GetApplicationFiles([]resources.AppFileResource{file1, file2, file3, file4})
			Expect(matchedFiles).To(Equal([]resources.AppFileResource{file3, file4}))
			Expect(err).NotTo(HaveOccurred())
		})

		It("excludes files that were in the response but not in the request", func() {
			setupTestServer(matchResourceRequestImbalanced)
			matchedFiles, err := repo.GetApplicationFiles([]resources.AppFileResource{file1, file4})
			Expect(matchedFiles).To(Equal([]resources.AppFileResource{file4}))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var matchedResources = testnet.RemoveWhiteSpaceFromBody(`[
	{
        "sha1": "d9c3a51de5c89c11331d3b90b972789f1a14699a",
        "size": 59
    },
    {
        "sha1": "345f999aef9070fb9a608e65cf221b7038156b6d",
        "size": 229
    }
]`)

var unmatchedResources = testnet.RemoveWhiteSpaceFromBody(`[
	{
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51,
        "fn": "app.rb",
				"mode":""
    },
    {
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70,
        "fn": "config.ru",
				"mode":""
    }
]`)

func uploadApplicationRequest(zipCheck func(*zip.Reader)) testnet.TestRequest {
	return testapi.NewCloudControllerTestRequest(testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/apps/my-cool-app-guid/bits",
		Matcher: uploadBodyMatcher(zipCheck),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body: `
{
	"metadata":{
		"guid": "my-job-guid",
		"url": "/v2/jobs/my-job-guid"
	}
}
	`},
	})
}

var matchResourceRequest = testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(testnet.RemoveWhiteSpaceFromBody(`[
	{
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70
    },
    {
        "sha1": "d9c3a51de5c89c11331d3b90b972789f1a14699a",
        "size": 59
    },
    {
        "sha1": "345f999aef9070fb9a608e65cf221b7038156b6d",
        "size": 229
    }
]`)),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   matchedResources,
	},
}

var matchResourceRequestImbalanced = testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(testnet.RemoveWhiteSpaceFromBody(`[
	{
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "sha1": "345f999aef9070fb9a608e65cf221b7038156b6d",
        "size": 229
    }
]`)),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   matchedResources,
	},
}

var defaultZipCheck = func(zipReader *zip.Reader) {
	Expect(len(zipReader.File)).To(Equal(2), "Wrong number of files in zip")

	var expectedPermissionBits os.FileMode
	if runtime.GOOS == "windows" {
		expectedPermissionBits = 0111
	} else {
		expectedPermissionBits = 0755
	}

	Expect(zipReader.File[0].Name).To(Equal("app.rb"))
	Expect(executableBits(zipReader.File[0].Mode())).To(Equal(executableBits(expectedPermissionBits)))

nextFile:
	for _, f := range zipReader.File {
		for _, expected := range expectedApplicationContent {
			if f.Name == expected {
				continue nextFile
			}
		}
		Fail("Expected " + f.Name + " but did not find it")
	}
}

var defaultRequests = []testnet.TestRequest{
	uploadApplicationRequest(defaultZipCheck),
	createProgressEndpoint("running"),
	createProgressEndpoint("finished"),
}

var expectedApplicationContent = []string{"app.rb", "config.ru"}

const maxMultipartResponseSizeInBytes = 4096

func uploadBodyMatcher(zipChecks func(zipReader *zip.Reader)) func(*http.Request) {
	return func(request *http.Request) {
		defer GinkgoRecover()
		err := request.ParseMultipartForm(maxMultipartResponseSizeInBytes)
		if err != nil {
			Fail(fmt.Sprintf("Failed parsing multipart form %v", err))
			return
		}
		defer request.MultipartForm.RemoveAll()

		Expect(len(request.MultipartForm.Value)).To(Equal(1), "Should have 1 value")
		valuePart, ok := request.MultipartForm.Value["resources"]
		Expect(ok).To(BeTrue(), "Resource manifest not present")
		Expect(len(valuePart)).To(Equal(1), "Wrong number of values")

		resourceManifest := valuePart[0]
		chompedResourceManifest := strings.Replace(resourceManifest, "\n", "", -1)
		Expect(chompedResourceManifest).To(Equal(unmatchedResources), "Resources do not match")

		Expect(len(request.MultipartForm.File)).To(Equal(1), "Wrong number of files")

		fileHeaders, ok := request.MultipartForm.File["application"]
		Expect(ok).To(BeTrue(), "Application file part not present")
		Expect(len(fileHeaders)).To(Equal(1), "Wrong number of files")

		applicationFile := fileHeaders[0]
		Expect(applicationFile.Filename).To(Equal("application.zip"), "Wrong file name")

		file, err := applicationFile.Open()
		if err != nil {
			Fail(fmt.Sprintf("Cannot get multipart file %v", err.Error()))
			return
		}

		length, err := strconv.ParseInt(applicationFile.Header.Get("content-length"), 10, 64)
		if err != nil {
			Fail(fmt.Sprintf("Cannot convert content-length to int %v", err.Error()))
			return
		}

		if zipChecks != nil {
			zipReader, err := zip.NewReader(file, length)
			if err != nil {
				Fail(fmt.Sprintf("Error reading zip content %v", err.Error()))
				return
			}

			zipChecks(zipReader)
		}
	}
}

func createProgressEndpoint(status string) (req testnet.TestRequest) {
	body := fmt.Sprintf(`
	{
		"entity":{
			"status":"%s"
		}
	}`, status)

	req.Method = "GET"
	req.Path = "/v2/jobs/my-job-guid"
	req.Response = testnet.TestResponse{
		Status: http.StatusCreated,
		Body:   body,
	}

	return
}

var matchExcludedResourceRequest = testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(testnet.RemoveWhiteSpaceFromBody(`[
    {
        "fn": ".svn",
        "sha1": "0",
        "size": 0
    },
    {
        "fn": ".svn/test",
        "sha1": "456b1d3f7cfbadc66d390de79cbbb6e6a10662da",
        "size": 12
    },
    {
        "fn": "_darcs",
        "sha1": "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
        "size": 4
    }
]`)),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   matchedResources,
	},
}

func executableBits(mode os.FileMode) os.FileMode {
	return mode & 0111
}
