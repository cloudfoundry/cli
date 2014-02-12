package api_test

import (
	"archive/zip"
	"cf"
	. "cf/api"
	"cf/models"
	"cf/net"
	"fileutils"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
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

var permissionsToSet os.FileMode
var expectedPermissionBits os.FileMode

func init() {
	permissionsToSet = 0467
	fileutils.TempFile("permissionedFile", func(file *os.File, err error) {
		if err != nil {
			panic("could not create tmp file")
		}

		fileInfo, err := file.Stat()
		if err != nil {
			panic("could not stat tmp file")
		}

		expectedPermissionBits = fileInfo.Mode()
		if runtime.GOOS != "windows" {
			expectedPermissionBits |= permissionsToSet & 0111
		}
	})
}

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

var uploadBodyMatcher = func(t mr.TestingT, request *http.Request) {
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
	Expect(zipReader.File[0].Mode()).To(Equal(os.FileMode(expectedPermissionBits)))

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

func testUploadApp(dir string, requests []testnet.TestRequest) (app models.Application, apiResponse net.ApiResponse) {
	ts, handler := testnet.NewTLSServer(GinkgoT(), requests)
	defer ts.Close()

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	gateway.PollingThrottle = time.Duration(0)
	zipper := cf.ApplicationZipper{}
	repo := NewCloudControllerApplicationBitsRepository(configRepo, gateway, zipper)

	var (
		reportedPath                          string
		reportedFileCount, reportedUploadSize uint64
	)
	apiResponse = repo.UploadApp("my-cool-app-guid", dir, func(path string, uploadSize, fileCount uint64) {
		reportedPath = path
		reportedUploadSize = uploadSize
		reportedFileCount = fileCount
	})

	Expect(reportedPath).To(Equal(dir))
	Expect(reportedFileCount).To(Equal(uint64(len(expectedApplicationContent))))
	Expect(reportedUploadSize).To(Equal(uint64(759)))
	Expect(handler.AllRequestsCalled()).To(BeTrue())

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUploadWithInvalidDirectory", func() {
		config := testconfig.NewRepository()
		gateway := net.NewCloudControllerGateway()
		zipper := &cf.ApplicationZipper{}

		repo := NewCloudControllerApplicationBitsRepository(config, gateway, zipper)

		apiResponse := repo.UploadApp("app-guid", "/foo/bar", func(path string, uploadSize, fileCount uint64) {})
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(ContainSubstring(filepath.Join("foo", "bar")))
	})

	It("TestUploadApp", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-app")
		err = os.Chmod(filepath.Join(dir, "Gemfile"), permissionsToSet)

		Expect(err).NotTo(HaveOccurred())

		_, apiResponse := testUploadApp(dir, defaultRequests)
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestCreateUploadDirWithAZipFile", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-app.zip")

		_, apiResponse := testUploadApp(dir, defaultRequests)
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestCreateUploadDirWithAZipLikeFile", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-app.azip")

		_, apiResponse := testUploadApp(dir, defaultRequests)
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestUploadAppFailsWhilePushingBits", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-app")

		requests := []testnet.TestRequest{
			matchResourceRequest,
			uploadApplicationRequest,
			createProgressEndpoint("running"),
			createProgressEndpoint("failed"),
		}
		_, apiResponse := testUploadApp(dir, requests)
		Expect(apiResponse.IsSuccessful()).To(BeFalse())
	})
})
