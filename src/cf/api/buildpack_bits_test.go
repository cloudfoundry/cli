package api_test

import (
	"archive/zip"
	"cf"
	. "cf/api"
	"cf/configuration"
	"cf/models"
	"cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	testconfig "testhelpers/configuration"
	testnet "testhelpers/net"
)

var _ = Describe("BuildpackBitsRepository", func() {
	var (
		buildpacksDir string
		configRepo    configuration.Repository
		repo          CloudControllerBuildpackBitsRepository
		buildpack     models.Buildpack
	)

	BeforeEach(func() {
		gateway := net.NewCloudControllerGateway()
		pwd, _ := os.Getwd()

		buildpacksDir = filepath.Join(pwd, "../../fixtures/buildpacks")
		configRepo = testconfig.NewRepositoryWithDefaults()
		repo = NewCloudControllerBuildpackBitsRepository(configRepo, gateway, cf.ApplicationZipper{})
		buildpack = models.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}
	})

	Describe("#UploadBuildpack", func() {
		It("fails to upload a buildpack with an invalid directory", func() {
			apiResponse := repo.UploadBuildpack(buildpack, "/foo/bar")
			Expect(apiResponse.IsNotSuccessful()).To(BeTrue())
			Expect(apiResponse.Message).To(ContainSubstring("Error opening buildpack file"))
		})

		It("uploads a valid buildpack directory", func() {
			buildpackPath := filepath.Join(buildpacksDir, "example-buildpack")

			os.Chmod(filepath.Join(buildpackPath, "bin/compile"), 0755)
			os.Chmod(filepath.Join(buildpackPath, "bin/detect"), 0755)
			err := os.Chmod(filepath.Join(buildpackPath, "bin/release"), 0755)
			Expect(err).NotTo(HaveOccurred())

			ts, handler := testnet.NewTLSServer([]testnet.TestRequest{
				uploadBuildpackRequest(buildpackPath),
			})
			defer ts.Close()
			configRepo.SetApiEndpoint(ts.URL)

			apiResponse := repo.UploadBuildpack(buildpack, buildpackPath)
			Expect(handler.AllRequestsCalled()).To(BeTrue())
			Expect(apiResponse.IsSuccessful()).To(BeTrue())
		})

		It("uploads a valid zipped buildpack", func() {
			buildpackPath := filepath.Join(buildpacksDir, "example-buildpack.zip")

			ts, handler := testnet.NewTLSServer([]testnet.TestRequest{
				uploadBuildpackRequest(buildpackPath),
			})
			defer ts.Close()

			configRepo.SetApiEndpoint(ts.URL)

			apiResponse := repo.UploadBuildpack(buildpack, buildpackPath)
			Expect(handler.AllRequestsCalled()).To(BeTrue())
			Expect(apiResponse.IsSuccessful()).To(BeTrue())
		})

		Describe("when the buildpack is wrapped in an extra top-level directory", func() {
			It("uploads a zip file containing only the actual buildpack", func() {
				buildpackPath := filepath.Join(buildpacksDir, "example-buildpack-in-dir.zip")

				ts, handler := testnet.NewTLSServer([]testnet.TestRequest{
					uploadBuildpackRequest(buildpackPath),
				})
				defer ts.Close()

				configRepo.SetApiEndpoint(ts.URL)

				apiResponse := repo.UploadBuildpack(buildpack, buildpackPath)
				Expect(handler.AllRequestsCalled()).To(BeTrue())
				Expect(apiResponse.IsSuccessful()).To(BeTrue())
			})
		})

		Describe("when given the URL of a buildpack", func() {
			var handler *testnet.TestHandler
			var apiServer *httptest.Server

			BeforeEach(func() {
				apiServer, handler = testnet.NewTLSServer([]testnet.TestRequest{
					uploadBuildpackRequest("example-buildpack.zip"),
				})
				configRepo.SetApiEndpoint(apiServer.URL)
			})

			AfterEach(func() {
				apiServer.Close()
			})

			var buildpackFileServerHandler = func(buildpackName string) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					Expect(request.URL.Path).To(Equal("/place/example-buildpack.zip"))
					f, err := os.Open(filepath.Join(buildpacksDir, buildpackName))
					Expect(err).NotTo(HaveOccurred())
					io.Copy(writer, f)
				}
			}

			It("uploads the file over HTTP", func() {
				fileServer := httptest.NewServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				apiResponse := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")
				Expect(handler.AllRequestsCalled()).To(BeTrue())
				Expect(apiResponse.IsSuccessful()).To(BeTrue())
			})

			It("uploads the file over HTTPS", func() {
				fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				apiResponse := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")
				Expect(handler.AllRequestsCalled()).To(BeTrue())
				Expect(apiResponse.IsSuccessful()).To(BeTrue())
			})

			Describe("when the buildpack is wrapped in an extra top-level directory", func() {
				It("uploads a zip file containing only the actual buildpack", func() {
					fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack-in-dir.zip"))
					defer fileServer.Close()

					apiResponse := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")
					Expect(handler.AllRequestsCalled()).To(BeTrue())
					Expect(apiResponse.IsSuccessful()).To(BeTrue())
				})
			})

			It("returns an unsuccessful response when the server cannot be reached", func() {
				apiResponse := repo.UploadBuildpack(buildpack, "https://domain.bad-domain:223453/no-place/example-buildpack.zip")
				Expect(handler.AllRequestsCalled()).To(BeFalse())
				Expect(apiResponse.IsSuccessful()).To(BeFalse())
			})
		})
	})
})

func uploadBuildpackRequest(filename string) testnet.TestRequest {
	return testnet.TestRequest{
		Method: "PUT",
		Path:   "/v2/buildpacks/my-cool-buildpack-guid/bits",
		Response: testnet.TestResponse{
			Status: http.StatusCreated,
			Body:   `{ "metadata":{ "guid": "my-job-guid" } }`,
		},
		Matcher: func(request *http.Request) {
			err := request.ParseMultipartForm(4096)
			defer request.MultipartForm.RemoveAll()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(request.MultipartForm.Value)).To(Equal(0))
			Expect(len(request.MultipartForm.File)).To(Equal(1))

			files, ok := request.MultipartForm.File["buildpack"]
			Expect(ok).To(BeTrue(), "Buildpack file part not present")
			Expect(len(files)).To(Equal(1), "Wrong number of files")

			buildpackFile := files[0]
			Expect(buildpackFile.Filename).To(Equal(filepath.Base(filename)), "Wrong file name")

			file, err := buildpackFile.Open()
			Expect(err).NotTo(HaveOccurred())

			zipReader, err := zip.NewReader(file, 4096)
			Expect(err).NotTo(HaveOccurred())

			actualFileNames := []string{}
			actualFileContents := []string{}
			for _, f := range zipReader.File {
				actualFileNames = append(actualFileNames, f.Name)
				c, _ := f.Open()
				content, _ := ioutil.ReadAll(c)
				actualFileContents = append(actualFileContents, string(content))
			}
			sort.Strings(actualFileNames)

			Expect(actualFileNames).To(Equal([]string{
				"bin/compile",
				"bin/detect",
				"bin/release",
				"lib/helper",
			}))
			Expect(actualFileContents).To(Equal([]string{
				"the-compile-script\n",
				"the-detect-script\n",
				"the-release-script\n",
				"the-helper-script\n",
			}))

			Expect(zipReader.File[1].Mode()).To(Equal(os.FileMode(0755)))
			Expect(zipReader.File[0].Mode()).To(Equal(os.FileMode(0755)))
			Expect(zipReader.File[2].Mode()).To(Equal(os.FileMode(0755)))
		},
	}
}
