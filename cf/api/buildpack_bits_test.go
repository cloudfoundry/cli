package api_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testnet "code.cloudfoundry.org/cli/util/testhelpers/net"

	. "code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildpackBitsRepository", func() {
	var (
		buildpacksDir     string
		configRepo        coreconfig.Repository
		repo              CloudControllerBuildpackBitsRepository
		buildpack         models.Buildpack
		testServer        *httptest.Server
		testServerHandler *testnet.TestHandler
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		pwd, _ := os.Getwd()

		buildpacksDir = filepath.Join(pwd, "../../fixtures/buildpacks")
		repo = NewCloudControllerBuildpackBitsRepository(configRepo, gateway, appfiles.ApplicationZipper{})
		buildpack = models.Buildpack{Name: "my-cool-buildpack", GUID: "my-cool-buildpack-guid"}

		testServer, testServerHandler = testnet.NewServer([]testnet.TestRequest{uploadBuildpackRequest()})
		configRepo.SetAPIEndpoint(testServer.URL)
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("CreateBuildpackZipFile", func() {

		Context("when buildpack path is a directory", func() {
			It("returns an error with an invalid directory", func() {
				_, _, err := repo.CreateBuildpackZipFile("/foo/bar")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Error opening buildpack file"))
			})

			It("does not return an error when it is a valid directory", func() {
				buildpackPath := filepath.Join(buildpacksDir, "example-buildpack")
				zipFile, zipFileName, err := repo.CreateBuildpackZipFile(buildpackPath)

				Expect(zipFileName).To(Equal("example-buildpack.zip"))
				Expect(zipFile).NotTo(BeNil())
				Expect(zipFile.Name()).To(ContainSubstring("buildpack-upload"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when buildpack path is a file", func() {
			It("returns an error", func() {
				_, _, err := repo.CreateBuildpackZipFile(filepath.Join(buildpacksDir, "file"))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a valid zip file"))
			})
		})

		Context("when buildpack path is a URL", func() {
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

					_, _, apiErr := repo.CreateBuildpackZipFile(fileServer.URL + "/place/bad-buildpack.zip")

					Expect(apiErr).To(HaveOccurred())
				})
			})

			It("download and create zip file over HTTP", func() {
				fileServer := httptest.NewServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				zipFile, zipFileName, apiErr := repo.CreateBuildpackZipFile(fileServer.URL + "/place/example-buildpack.zip")

				Expect(zipFileName).To(Equal("example-buildpack.zip"))
				Expect(zipFile).NotTo(BeNil())
				Expect(zipFile.Name()).To(ContainSubstring("buildpack-upload"))
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("download and create zip file over HTTPS", func() {
				fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack.zip"))
				defer fileServer.Close()

				repo.TrustedCerts = fileServer.TLS.Certificates
				zipFile, zipFileName, apiErr := repo.CreateBuildpackZipFile(fileServer.URL + "/place/example-buildpack.zip")

				Expect(zipFileName).To(Equal("example-buildpack.zip"))
				Expect(zipFile).NotTo(BeNil())
				Expect(zipFile.Name()).To(ContainSubstring("buildpack-upload"))
				Expect(apiErr).NotTo(HaveOccurred())
			})

			It("fails when the server's SSL cert cannot be verified", func() {
				fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack.zip"))
				fileServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
				defer fileServer.Close()

				_, _, apiErr := repo.CreateBuildpackZipFile(fileServer.URL + "/place/example-buildpack.zip")

				Expect(apiErr).To(HaveOccurred())
			})

			Context("when the buildpack is wrapped in an extra top-level directory", func() {
				It("uploads a zip file containing only the actual buildpack", func() {
					fileServer := httptest.NewTLSServer(buildpackFileServerHandler("example-buildpack-in-dir.zip"))
					defer fileServer.Close()

					repo.TrustedCerts = fileServer.TLS.Certificates
					zipFile, zipFileName, apiErr := repo.CreateBuildpackZipFile(fileServer.URL + "/place/example-buildpack-in-dir.zip")

					Expect(zipFileName).To(Equal("example-buildpack-in-dir.zip"))
					Expect(zipFile).NotTo(BeNil())
					Expect(zipFile.Name()).To(ContainSubstring("buildpack-upload"))
					Expect(apiErr).NotTo(HaveOccurred())
				})
			})

			It("returns an unsuccessful response when the server cannot be reached", func() {
				_, _, apiErr := repo.CreateBuildpackZipFile("https://domain.bad-domain:223453/no-place/example-buildpack.zip")

				Expect(apiErr).To(HaveOccurred())
			})
		})
	})

	Describe("UploadBuildpack", func() {
		var (
			zipFileName string
			zipFile     *os.File
			err         error
		)

		JustBeforeEach(func() {
			buildpackPath := filepath.Join(buildpacksDir, zipFileName)
			zipFile, _, err = repo.CreateBuildpackZipFile(buildpackPath)

			Expect(zipFile).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when it is a valid zipped buildpack", func() {
			BeforeEach(func() {
				zipFileName = "example-buildpack.zip"
			})

			It("uploads the buildpack", func() {

				apiErr := repo.UploadBuildpack(buildpack, zipFile, zipFileName)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(testServerHandler).To(HaveAllRequestsCalled())
			})
		})

		Describe("when the buildpack is wrapped in an extra top-level directory", func() {
			BeforeEach(func() {
				zipFileName = "example-buildpack-in-dir.zip"
			})

			It("uploads a zip file containing only the actual buildpack", func() {
				apiErr := repo.UploadBuildpack(buildpack, zipFile, zipFileName)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(testServerHandler).To(HaveAllRequestsCalled())
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

			Expect(buildpackFile.Filename).To(ContainSubstring(".zip"))

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
