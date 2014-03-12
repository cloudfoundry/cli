package api_test

import (
	"archive/zip"
	"cf/app_files"
	. "cf/api"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	testapi "testhelpers/api"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
	"time"
)

var _ = Describe("CloudControllerApplicationBitsRepository", func() {
	var fixturesDir string

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = filepath.Join(cwd, "../../fixtures/applications")
	})

	It("TestUploadWithInvalidDirectory", func() {
		config := testconfig.NewRepository()
		gateway := net.NewCloudControllerGateway(config)
		zipper := &app_files.ApplicationZipper{}

		repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

		apiErr := repo.UploadApp("app-guid", "/foo/bar", func(path string, uploadSize, fileCount uint64) {})
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Error()).To(ContainSubstring(filepath.Join("foo", "bar")))
	})

	Context("uploading a directory", func() {
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
			_, apiErr := testUploadApp(appPath, defaultRequests)
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("returns a failure when uploading bits fails", func() {
			requests := []testnet.TestRequest{
				matchResourceRequest,
				uploadApplicationRequest,
				createProgressEndpoint("running"),
				createProgressEndpoint("failed"),
			}
			_, apiErr := testUploadApp(appPath, requests)
			Expect(apiErr).To(HaveOccurred())
		})
	})

	It("uploads zip files", func() {
		_, apiErr := testUploadApp(filepath.Join(fixturesDir, "example-app.zip"), defaultRequests)
		Expect(apiErr).NotTo(HaveOccurred())
	})

	It("uploads zip files with non-standard names", func() {
		_, apiErr := testUploadApp(filepath.Join(fixturesDir, "example-app.azip"), defaultRequests)
		Expect(apiErr).NotTo(HaveOccurred())
	})
})

var expectedResources = testnet.RemoveWhiteSpaceFromBody(`[
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
    },
    {
        "fn": "manifest.yml",
        "sha1": "19b5b4225dc64da3213b1ffaa1e1920ee5faf36c",
        "size": 111
    }
]`)

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
	Method:  "PUT",
	Path:    "/v2/resource_match",
	Matcher: testnet.RequestBodyMatcher(expectedResources),
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

var expectedApplicationContent = []string{"Gemfile", "Gemfile.lock", "manifest.yml"}

var uploadBodyMatcher = func(request *http.Request) {
	err := request.ParseMultipartForm(4096)
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

	Expect(len(zipReader.File)).To(Equal(3), "Wrong number of files in zip")

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

func testUploadApp(dir string, requests []testnet.TestRequest) (app models.Application, apiErr errors.Error) {
	ts, handler := testnet.NewServer(requests)
	defer ts.Close()

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway(configRepo)
	gateway.PollingThrottle = time.Duration(0)
	zipper := app_files.ApplicationZipper{}
	repo := NewCloudControllerApplicationBitsRepository(configRepo, gateway, zipper)

	var (
		reportedPath                          string
		reportedFileCount, reportedUploadSize uint64
	)
	apiErr = repo.UploadApp("my-cool-app-guid", dir, func(path string, uploadSize, fileCount uint64) {
		reportedPath = path
		reportedUploadSize = uploadSize
		reportedFileCount = fileCount
	})

	Expect(reportedPath).To(Equal(dir))
	Expect(reportedFileCount).To(Equal(uint64(len(expectedApplicationContent))))
	Expect(reportedUploadSize).To(Equal(uint64(759)))
	Expect(handler).To(testnet.HaveAllRequestsCalled())

	return
}
