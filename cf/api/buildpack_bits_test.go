package api_test

import (
	"archive/zip"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testnet "github.com/cloudfoundry/cli/testhelpers/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

var _ = Describe("BuildpackBitsRepository", func() {
	var (
		buildpacksDir     string
		configRepo        configuration.Repository
		repo              CloudControllerBuildpackBitsRepository
		buildpack         models.Buildpack
		testServer        *httptest.Server
		testServerHandler *testnet.TestHandler
	)

	BeforeEach(func() {
		gateway := net.NewCloudControllerGateway(configRepo)
		pwd, _ := os.Getwd()

		buildpacksDir = filepath.Join(pwd, "../../fixtures/buildpacks")
		configRepo = testconfig.NewRepositoryWithDefaults()
		repo = NewCloudControllerBuildpackBitsRepository(configRepo, gateway, app_files.ApplicationZipper{})
		buildpack = models.Buildpack{Name: "my-cool-buildpack", Guid: "my-cool-buildpack-guid"}

		testServer, testServerHandler = testnet.NewServer([]testnet.TestRequest{uploadBuildpackRequest()})
		configRepo.SetApiEndpoint(testServer.URL)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("#UploadBuildpack", func() {
		It("fails to upload a buildpack with an invalid directory", func() {
			apiErr := repo.UploadBuildpack(buildpack, "/foo/bar")
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Error()).To(ContainSubstring("Error opening buildpack file"))
		})

		It("uploads a valid buildpack directory", func() {
			buildpackPath := filepath.Join(buildpacksDir, "example-buildpack")

			os.Chmod(filepath.Join(buildpackPath, "bin/compile"), 0755)
			os.Chmod(filepath.Join(buildpackPath, "bin/detect"), 0755)
			err := os.Chmod(filepath.Join(buildpackPath, "bin/release"), 0755)
			Expect(err).NotTo(HaveOccurred())

			apiErr := repo.UploadBuildpack(buildpack, buildpackPath)
			Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		It("uploads a valid zipped buildpack", func() {
			buildpackPath := filepath.Join(buildpacksDir, "example-buildpack.zip")

			apiErr := repo.UploadBuildpack(buildpack, buildpackPath)
			Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
			Expect(apiErr).NotTo(HaveOccurred())
		})

		Describe("when the buildpack is wrapped in an extra top-level directory", func() {
			It("uploads a zip file containing only the actual buildpack", func() {
				buildpackPath := filepath.Join(buildpacksDir, "example-buildpack-in-dir.zip")

				apiErr := repo.UploadBuildpack(buildpack, buildpackPath)
				Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})
		})

		Describe("when given the URL of a buildpack", func() {
			var buildpackFileServerHandler = func(buildpackName string) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					Expect(request.URL.Path).To(Equal(fmt.Sprintf("/place/%s", buildpackName)))
					f, err := os.Open(filepath.Join(buildpacksDir, buildpackName))
					Expect(err).NotTo(HaveOccurred())
					io.Copy(writer, f)
				}
			}

			Context("when the downloaded resource is not a valid zip file", func() {
				It("fails gracefully", func() {
					fileServer := httptest.NewServer(buildpackFileServerHandler("bad-buildpack.zip"))
					defer fileServer.Close()

					apiErr := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/bad-buildpack.zip")
					Expect(testServerHandler).NotTo(testnet.HaveAllRequestsCalled())
					Expect(apiErr).To(HaveOccurred())
				})
			})

			It("uploads the file over HTTP", func() {
				fileServer := httptest.NewServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				apiErr := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")

				Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("uploads the file over HTTPS", func() {
				fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				repo.TrustedCerts = fileServer.TLS.Certificates
				apiErr := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")

				Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("fails when the server's SSL cert cannot be verified", func() {
				fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				apiErr := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack.zip")

				Expect(testServerHandler).NotTo(testnet.HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
			})

			Describe("when the buildpack is wrapped in an extra top-level directory", func() {
				It("uploads a zip file containing only the actual buildpack", func() {
					fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack-in-dir.zip"))
					defer fileServer.Close()

					repo.TrustedCerts = fileServer.TLS.Certificates
					apiErr := repo.UploadBuildpack(buildpack, fileServer.URL+"/place/example-buildpack-in-dir.zip")

					Expect(testServerHandler).To(testnet.HaveAllRequestsCalled())
					Expect(apiErr).NotTo(HaveOccurred())
				})
			})

			It("returns an unsuccessful response when the server cannot be reached", func() {
				apiErr := repo.UploadBuildpack(buildpack, "https://domain.bad-domain:223453/no-place/example-buildpack.zip")
				Expect(testServerHandler).NotTo(testnet.HaveAllRequestsCalled())
				Expect(apiErr).To(HaveOccurred())
			})
		})
	})
})

func uploadBuildpackRequest() testnet.TestRequest {
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
				"bin/",
				"bin/compile",
				"bin/detect",
				"bin/release",
				"lib/",
				"lib/helper",
			}))
			Expect(actualFileContents).To(Equal([]string{
				"",
				"the-compile-script\n",
				"the-detect-script\n",
				"the-release-script\n",
				"",
				"the-helper-script\n",
			}))

			if runtime.GOOS != "windows" {
				for i := 1; i < 4; i++ {
					Expect(zipReader.File[i].Mode()).To(Equal(os.FileMode(0755)))
				}
			}
		},
	}
}
