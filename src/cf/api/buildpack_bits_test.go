package api_test

import (
	"archive/zip"
	"cf"
	. "cf/api"
	"cf/models"
	"cf/net"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	"net/http"
	"os"
	"path/filepath"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("BuildpackBitsRepository", func() {
	It("TestUploadBuildpackWithInvalidDirectory", func() {
		config := testconfig.NewRepository()
		gateway := net.NewCloudControllerGateway()

		repo := NewCloudControllerBuildpackBitsRepository(config, gateway, cf.ApplicationZipper{})
		buildpack := models.Buildpack{}

		apiResponse := repo.UploadBuildpack(buildpack, "/foo/bar")
		Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
		Expect(apiResponse.Message).To(ContainSubstring("Invalid buildpack"))
	})

	It("TestUploadBuildpack", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-buildpack")
		err = os.Chmod(filepath.Join(dir, "detect"), 0666)
		Expect(err).NotTo(HaveOccurred())

		_, apiResponse := testUploadBuildpack(mr.T(), dir, []testnet.TestRequest{
			uploadBuildpackRequest(dir),
		})
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})

	It("TestUploadBuildpackWithAZipFile", func() {
		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		dir = filepath.Join(dir, "../../fixtures/example-buildpack.zip")

		_, apiResponse := testUploadBuildpack(mr.T(), dir, []testnet.TestRequest{
			uploadBuildpackRequest(dir),
		})
		Expect(apiResponse.IsSuccessful()).To(BeTrue())
	})
})

func uploadBuildpackRequest(filename string) testnet.TestRequest {
	return testnet.TestRequest{
		Method:  "PUT",
		Path:    "/v2/buildpacks/my-cool-buildpack-guid/bits",
		Matcher: uploadBuildpackBodyMatcher(filename),
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body: `
{
	"metadata":{
		"guid": "my-job-guid"
	}
}
	`},
	}
}

var expectedBuildpackContent = []string{"detect", "compile", "package"}

func uploadBuildpackBodyMatcher(pathToFile string) testnet.RequestMatcher {
	return func(t mr.TestingT, request *http.Request) {
		err := request.ParseMultipartForm(4096)
		if err != nil {
			Fail(fmt.Sprintf("Failed parsing multipart form: %s", err))
			return
		}
		defer request.MultipartForm.RemoveAll()

		Expect(len(request.MultipartForm.Value)).To(Equal(0), "Should have 0 values")
		Expect(len(request.MultipartForm.File)).To(Equal(1), "Wrong number of files")

		files, ok := request.MultipartForm.File["buildpack"]

		Expect(ok).To(BeTrue(), "Buildpack file part not present")
		Expect(len(files)).To(Equal(1), "Wrong number of files")

		buildpackFile := files[0]
		Expect(buildpackFile.Filename).To(Equal(filepath.Base(pathToFile)), "Wrong file name")

		file, err := buildpackFile.Open()
		if err != nil {
			Fail(fmt.Sprintf("Cannot get multipart file: %s", err.Error()))
			return
		}

		zipReader, err := zip.NewReader(file, 4096)
		if err != nil {
			Fail(fmt.Sprintf("Error reading zip content: %s", err.Error()))
		}

		Expect(len(zipReader.File)).To(Equal(3), "Wrong number of files in zip")
		Expect(zipReader.File[1].Mode()).To(Equal(os.FileMode(0666)))

	nextFile:
		for _, f := range zipReader.File {
			for _, expected := range expectedBuildpackContent {
				if f.Name == expected {
					continue nextFile
				}
			}
			Fail("Missing file: " + f.Name)
		}
	}
}

func testUploadBuildpack(t mr.TestingT, dir string, requests []testnet.TestRequest) (buildpack models.Buildpack, apiResponse net.ApiResponse) {
	ts, handler := testnet.NewTLSServer(t, requests)
	defer ts.Close()

	configRepo := testconfig.NewRepositoryWithDefaults()
	configRepo.SetApiEndpoint(ts.URL)
	gateway := net.NewCloudControllerGateway()
	repo := NewCloudControllerBuildpackBitsRepository(configRepo, gateway, cf.ApplicationZipper{})
	buildpack = models.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}

	apiResponse = repo.UploadBuildpack(buildpack, dir)
	Expect(handler.AllRequestsCalled()).To(BeTrue())
	return
}
