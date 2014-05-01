package api_test

import (
	"archive/zip"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/net"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("CloudControllerApplicationBitsRepository", func() {
	var (
		fixturesDir        string
		reportedFilePath   string
		reportedUploadSize uint64
		reportedFileCount  uint64
	)

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = filepath.Join(cwd, "../../fixtures/applications")
	})

	var testUploadApp = func(dir string, requests ...testnet.TestRequest) (apiErr error) {
		ts, handler := testnet.NewServer(requests)
		defer ts.Close()

		configRepo := testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(ts.URL)
		gateway := net.NewCloudControllerGateway(configRepo)
		gateway.PollingThrottle = time.Duration(0)
		zipper := app_files.ApplicationZipper{}
		repo := NewCloudControllerApplicationBitsRepository(configRepo, gateway, zipper)

		apiErr = repo.UploadApp("my-cool-app-guid", dir, func(path string, uploadSize, fileCount uint64) {
			reportedFilePath = path
			reportedUploadSize = uploadSize
			reportedFileCount = fileCount
		})

		Expect(handler).To(testnet.HaveAllRequestsCalled())
		return
	}

	It("uploads zip files", func() {
		appPath := filepath.Join(fixturesDir, "example-app.azip")
		apiErr := testUploadApp(appPath, defaultRequests...)

		Expect(apiErr).NotTo(HaveOccurred())
		Expect(reportedFilePath).To(Equal(appPath))
		Expect(reportedFileCount).To(Equal(uint64(len(expectedApplicationContent))))
		Expect(reportedUploadSize).To(Equal(uint64(532)))
	})

	It("returns an error when the directory to upload does not exist", func() {
		config := testconfig.NewRepository()
		gateway := net.NewCloudControllerGateway(config)
		zipper := &app_files.ApplicationZipper{}

		repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

		apiErr := repo.UploadApp("app-guid", "/foo/bar", func(_ string, _, _ uint64) {})
		Expect(apiErr).To(HaveOccurred())
		Expect(apiErr.Error()).To(ContainSubstring(filepath.Join("foo", "bar")))
	})

	Context("when uploading a directory", func() {
		var appPath string

		BeforeEach(func() {
			appPath = filepath.Join(fixturesDir, "example-app")

			// the executable bit is the only bit we care about here
			err := os.Chmod(filepath.Join(appPath, "Gemfile"), 0467)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Chmod(filepath.Join(appPath, "Gemfile"), 0666)
		})

		It("preserves the executable bits when uploading app files", func() {
			apiErr := testUploadApp(appPath, defaultRequests...)
			Expect(apiErr).NotTo(HaveOccurred())

			Expect(reportedFilePath).To(Equal(appPath))
			Expect(reportedFileCount).To(Equal(uint64(len(expectedApplicationContent))))
			Expect(reportedUploadSize).To(Equal(uint64(532)))
		})

		It("returns a failure when uploading bits fails", func() {
			apiErr := testUploadApp(appPath, []testnet.TestRequest{
				matchResourceRequest,
				testapi.NewCloudControllerTestRequest(testnet.TestRequest{
					Method:  "PUT",
					Path:    "/v2/apps/my-cool-app-guid/bits",
					Matcher: uploadBodyMatcher,
					Response: testnet.TestResponse{
						Status: http.StatusCreated,
						Body: `
					{
						"metadata":{
							"guid": "my-job-guid",
							"url": "/v2/jobs/my-job-guid"
						}
					}`,
					}}),
				createProgressEndpoint("running"),
				createProgressEndpoint("failed"),
			}...)

			Expect(apiErr).To(HaveOccurred())
		})

		Context("when there are no files to upload", func() {
			It("makes a request without a zipfile", func() {
				emptyDir := filepath.Join(fixturesDir, "empty-dir")
				err := testUploadApp(emptyDir,
					testnet.TestRequest{
						Method:  "PUT",
						Path:    "/v2/resource_match",
						Matcher: testnet.RequestBodyMatcher("[]"),
						Response: testnet.TestResponse{
							Status: http.StatusOK,
							Body:   "[]",
						},
					},
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
						}}),
					createProgressEndpoint("running"),
					createProgressEndpoint("finished"),
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(reportedFileCount).To(Equal(uint64(0)))
				Expect(reportedUploadSize).To(Equal(uint64(0)))
				Expect(reportedFilePath).To(Equal(emptyDir))
			})
		})
	})
})

var matchedResources = testnet.RemoveWhiteSpaceFromBody(`[
	{
        "fn": "app.rb",
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "fn": "config.ru",
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70
    }
]`)

var uploadApplicationRequest = testapi.NewCloudControllerTestRequest(testnet.TestRequest{
	Method:  "PUT",
	Path:    "/v2/apps/my-cool-app-guid/bits",
	Matcher: uploadBodyMatcher,
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

var matchResourceRequest = testnet.TestRequest{
	Method: "PUT",
	Path:   "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(testnet.RemoveWhiteSpaceFromBody(`[
    {
        "fn": "Gemfile",
        "sha1": "d9c3a51de5c89c11331d3b90b972789f1a14699a",
        "size": 59
    },
    {
        "fn": "Gemfile.lock",
        "sha1": "345f999aef9070fb9a608e65cf221b7038156b6d",
        "size": 229
    },
    {
        "fn": "app.rb",
        "sha1": "2474735f5163ba7612ef641f438f4b5bee00127b",
        "size": 51
    },
    {
        "fn": "config.ru",
        "sha1": "f097424ce1fa66c6cb9f5e8a18c317376ec12e05",
        "size": 70
    }
]`)),
	Response: testnet.TestResponse{
		Status: http.StatusOK,
		Body:   matchedResources,
	},
}

var defaultRequests = []testnet.TestRequest{
	matchResourceRequest,
	uploadApplicationRequest,
	createProgressEndpoint("running"),
	createProgressEndpoint("finished"),
}

var expectedApplicationContent = []string{"Gemfile", "Gemfile.lock"}

const maxMultipartResponseSizeInBytes = 4096

var uploadBodyMatcher = func(request *http.Request) {
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
	Expect(chompedResourceManifest).To(Equal(matchedResources), "Resources do not match")

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

	zipReader, err := zip.NewReader(file, length)
	if err != nil {
		Fail(fmt.Sprintf("Error reading zip content %v", err.Error()))
		return
	}

	Expect(len(zipReader.File)).To(Equal(2), "Wrong number of files in zip")

	var expectedPermissionBits os.FileMode
	if runtime.GOOS == "windows" {
		expectedPermissionBits = 0666
	} else {
		expectedPermissionBits = 0467
	}

	Expect(zipReader.File[0].Name).To(Equal("Gemfile"))
	Expect(executableBits(zipReader.File[0].Mode())).To(Equal(executableBits(expectedPermissionBits)))

nextFile:
	for _, f := range zipReader.File {
		for _, expected := range expectedApplicationContent {
			if f.Name == expected {
				continue nextFile
			}
		}
		Fail("Missing file: " + f.Name)
	}
}

func executableBits(mode os.FileMode) os.FileMode {
	return mode & 0111
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
