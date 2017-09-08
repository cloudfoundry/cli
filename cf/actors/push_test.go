package actors_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/actors/actorsfakes"
	"code.cloudfoundry.org/cli/cf/api/applicationbits/applicationbitsfakes"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/appfiles/appfilesfakes"
	"code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Push Actor", func() {
	var (
		appBitsRepo  *applicationbitsfakes.FakeApplicationBitsRepository
		appFiles     *appfilesfakes.FakeAppFiles
		fakezipper   *appfilesfakes.FakeZipper
		routeActor   *actorsfakes.FakeRouteActor
		actor        actors.PushActor
		fixturesDir  string
		appDir       string
		allFiles     []models.AppFileFields
		presentFiles []resources.AppFileResource
	)

	BeforeEach(func() {
		appBitsRepo = new(applicationbitsfakes.FakeApplicationBitsRepository)
		appFiles = new(appfilesfakes.FakeAppFiles)
		fakezipper = new(appfilesfakes.FakeZipper)
		routeActor = new(actorsfakes.FakeRouteActor)
		actor = actors.NewPushActor(appBitsRepo, fakezipper, appFiles, routeActor)
		fixturesDir = filepath.Join("..", "..", "fixtures", "applications")
		allFiles = []models.AppFileFields{
			{Path: "example-app/.cfignore"},
			{Path: "example-app/app.rb"},
			{Path: "example-app/config.ru"},
			{Path: "example-app/Gemfile"},
			{Path: "example-app/Gemfile.lock"},
			{Path: "example-app/ignore-me"},
			{Path: "example-app/manifest.yml"},
		}
	})

	Describe("GatherFiles", func() {
		var tmpDir string

		BeforeEach(func() {
			presentFiles = []resources.AppFileResource{
				{Path: "example-app/ignore-me"},
			}

			appDir = filepath.Join(fixturesDir, "example-app.zip")
			appBitsRepo.GetApplicationFilesReturns(presentFiles, nil)
			var err error
			tmpDir, err = ioutil.TempDir("", "gather-files")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		Context("when we cannot reach CC", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("error")
				appBitsRepo.GetApplicationFilesReturns(nil, expectedErr)
			})

			It("returns an error if we cannot reach the cc", func() {
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir, true)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when we cannot copy the app files", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("error")
				appFiles.CopyFilesReturns(expectedErr)
			})

			It("returns an error", func() {
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir, true)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when using .cfignore", func() {
			BeforeEach(func() {
				appDir = filepath.Join(fixturesDir, "exclude-a-default-cfignore")
				// Ignore app files for this test as .cfignore is not one of them
				appBitsRepo.GetApplicationFilesReturns(nil, nil)
			})

			It("copies the .cfignore file to the upload directory", func() {
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(tmpDir, ".cfignore"))
				Expect(os.IsNotExist(err)).To(BeFalse())
			})
		})

		It("returns files to upload with file mode unchanged on non-Windows platforms", func() {
			if runtime.GOOS == "windows" {
				Skip("This does not run on windows")
			}

			info, err := os.Lstat(filepath.Join(fixturesDir, "example-app/ignore-me"))
			Expect(err).NotTo(HaveOccurred())

			expectedFileMode := fmt.Sprintf("%#o", info.Mode())

			actualFiles, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
			Expect(err).NotTo(HaveOccurred())

			expectedFiles := []resources.AppFileResource{
				{
					Path: "example-app/ignore-me",
					Mode: expectedFileMode,
				},
			}

			Expect(actualFiles).To(Equal(expectedFiles))
		})

		It("returns files to upload with file mode always being executable on Windows platforms", func() {
			if runtime.GOOS != "windows" {
				Skip("This runs only on windows")
			}

			info, err := os.Lstat(filepath.Join(fixturesDir, "example-app/ignore-me"))
			Expect(err).NotTo(HaveOccurred())

			expectedFileMode := fmt.Sprintf("%#o", info.Mode()|0700)

			actualFiles, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
			Expect(err).NotTo(HaveOccurred())

			expectedFiles := []resources.AppFileResource{
				{
					Path: "example-app/ignore-me",
					Mode: expectedFileMode,
				},
			}

			Expect(actualFiles).To(Equal(expectedFiles))
		})

		Context("when there are no remote files", func() {
			BeforeEach(func() {
				appBitsRepo.GetApplicationFilesReturns([]resources.AppFileResource{}, nil)
			})

			It("returns true for hasFileToUpload", func() {
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFileToUpload).To(BeTrue())
			})

			It("copies all local files to the upload dir", func() {
				expectedFiles := []models.AppFileFields{
					{Path: "example-app/.cfignore"},
					{Path: "example-app/app.rb"},
					{Path: "example-app/config.ru"},
					{Path: "example-app/Gemfile"},
					{Path: "example-app/Gemfile.lock"},
					{Path: "example-app/ignore-me"},
					{Path: "example-app/manifest.yml"},
				}
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, fromDir, toDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(Equal(expectedFiles))
				Expect(fromDir).To(Equal(fixturesDir))
				Expect(toDir).To(Equal(tmpDir))
			})
		})

		Context("when there are local files that aren't matched", func() {
			var remoteFiles []resources.AppFileResource

			BeforeEach(func() {
				remoteFiles = []resources.AppFileResource{
					{Path: "example-app/manifest.yml"},
				}

				appBitsRepo.GetApplicationFilesReturns(remoteFiles, nil)
			})

			It("returns true for hasFileToUpload", func() {
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFileToUpload).To(BeTrue())
			})

			It("copies unmatched local files to the upload dir", func() {
				expectedFiles := []models.AppFileFields{
					{Path: "example-app/.cfignore"},
					{Path: "example-app/app.rb"},
					{Path: "example-app/config.ru"},
					{Path: "example-app/Gemfile"},
					{Path: "example-app/Gemfile.lock"},
					{Path: "example-app/ignore-me"},
				}
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, fromDir, toDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(Equal(expectedFiles))
				Expect(fromDir).To(Equal(fixturesDir))
				Expect(toDir).To(Equal(tmpDir))
			})
		})

		Context("when local and remote files are equivalent", func() {
			BeforeEach(func() {
				remoteFiles := []resources.AppFileResource{
					{Path: "example-app/.cfignore", Mode: "0644"},
					{Path: "example-app/app.rb", Mode: "0755"},
					{Path: "example-app/config.ru", Mode: "0644"},
					{Path: "example-app/Gemfile", Mode: "0644"},
					{Path: "example-app/Gemfile.lock", Mode: "0644"},
					{Path: "example-app/ignore-me", Mode: "0666"},
					{Path: "example-app/manifest.yml", Mode: "0644"},
				}

				appBitsRepo.GetApplicationFilesReturns(remoteFiles, nil)
			})

			It("returns false for hasFileToUpload", func() {
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFileToUpload).To(BeFalse())
			})

			It("copies nothing to the upload dir", func() {
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir, true)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, fromDir, toDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(BeEmpty())
				Expect(fromDir).To(Equal(fixturesDir))
				Expect(toDir).To(Equal(tmpDir))
			})
		})

		Context("when told not to use the remote cache", func() {
			It("does not use the remote cache", func() {
				actor.GatherFiles(allFiles, fixturesDir, tmpDir, false)
				Expect(appBitsRepo.GetApplicationFilesCallCount()).To(Equal(0))
			})
		})
	})

	Describe("UploadApp", func() {
		It("Simply delegates to the UploadApp function on the app bits repo, which is not worth testing", func() {})
	})

	Describe("ProcessPath", func() {
		var (
			wasCalled     bool
			wasCalledWith string
		)

		BeforeEach(func() {
			zipper := &appfiles.ApplicationZipper{}
			actor = actors.NewPushActor(appBitsRepo, zipper, appFiles, routeActor)
		})

		Context("when given a zip file", func() {
			var zipFile string

			BeforeEach(func() {
				zipFile = filepath.Join(fixturesDir, "example-app.zip")
			})

			It("extracts the zip when given a zip file", func() {
				f := func(tempDir string) error {
					for _, file := range allFiles {
						actualFilePath := filepath.Join(tempDir, file.Path)
						_, err := os.Stat(actualFilePath)
						Expect(err).NotTo(HaveOccurred())
					}
					return nil
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls the provided function with the directory that it extracted to", func() {
				f := func(tempDir string) error {
					wasCalled = true
					wasCalledWith = tempDir
					return nil
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).NotTo(HaveOccurred())
				Expect(wasCalled).To(BeTrue())
				Expect(wasCalledWith).NotTo(Equal(zipFile))
			})

			It("cleans up the directory that it extracted to", func() {
				var tempDirWas string
				f := func(tempDir string) error {
					tempDirWas = tempDir
					return nil
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).NotTo(HaveOccurred())
				_, err = os.Stat(tempDirWas)
				Expect(err).To(HaveOccurred())
			})

			It("returns an error if the unzipping fails", func() {
				e := errors.New("some-error")
				fakezipper.UnzipReturns(e)
				fakezipper.IsZipFileReturns(true)
				actor = actors.NewPushActor(appBitsRepo, fakezipper, appFiles, routeActor)

				f := func(_ string) error {
					return nil
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).To(HaveOccurred())
			})
		})

		It("calls the provided function with the provided directory", func() {
			appDir = filepath.Join(fixturesDir, "example-app")
			f := func(tempDir string) error {
				wasCalled = true
				wasCalledWith = tempDir
				return nil
			}
			err := actor.ProcessPath(appDir, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalled).To(BeTrue())

			path, err := filepath.Abs(appDir)
			Expect(err).NotTo(HaveOccurred())

			path, err = filepath.EvalSymlinks(path)
			Expect(err).NotTo(HaveOccurred())

			Expect(wasCalledWith).To(Equal(path))
		})

		It("dereferences the symlink when given a symlink to an app dir", func() {
			if runtime.GOOS == "windows" {
				Skip("This should not run on Windows")
			}

			symlink := filepath.Join(fixturesDir, "example-app-symlink")
			expectedDir := filepath.Join(fixturesDir, "example-app") // example-app-symlink -> example-app
			f := func(dir string) error {
				wasCalled = true
				wasCalledWith = dir
				return nil
			}

			err := actor.ProcessPath(symlink, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalled).To(BeTrue())
			path, err := filepath.Abs(expectedDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalledWith).To(Equal(path))
		})

		It("calls the provided function with the provided absolute directory", func() {
			appDir = filepath.Join(fixturesDir, "example-app")
			absolutePath, err := filepath.Abs(appDir)
			Expect(err).NotTo(HaveOccurred())
			f := func(tempDir string) error {
				wasCalled = true
				wasCalledWith = tempDir
				return nil
			}
			err = actor.ProcessPath(absolutePath, f)
			Expect(err).NotTo(HaveOccurred())

			absolutePath, err = filepath.EvalSymlinks(absolutePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(wasCalled).To(BeTrue())
			Expect(wasCalledWith).To(Equal(absolutePath))
		})
	})

	Describe("ValidateAppParams", func() {
		var apps []models.AppParams

		Context("when HealthCheckType is not http", func() {
			Context("when HealthCheckHTTPEndpoint is provided", func() {
				BeforeEach(func() {
					healthCheckType := "port"
					endpoint := "/some-endpoint"
					apps = []models.AppParams{
						models.AppParams{
							HealthCheckType:         &healthCheckType,
							HealthCheckHTTPEndpoint: &endpoint,
						},
					}
				})
				It("displays error", func() {
					errs := actor.ValidateAppParams(apps)
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Error()).To(Equal("Health check type must be 'http' to set a health check HTTP endpoint."))
				})
			})
		})

		Context("when docker and buildpack is provided", func() {
			BeforeEach(func() {
				appName := "my-app"
				dockerImage := "some-image"
				buildpackURL := "some-build-pack.url"
				apps = []models.AppParams{
					models.AppParams{
						Name:         &appName,
						BuildpackURL: &buildpackURL,
						DockerImage:  &dockerImage,
					},
				}
			})
			It("displays error", func() {
				errs := actor.ValidateAppParams(apps)
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("Application my-app must not be configured with both 'buildpack' and 'docker'"))
			})
		})

		Context("when docker and path is provided", func() {
			BeforeEach(func() {
				appName := "my-app"
				dockerImage := "some-image"
				path := "some-path"
				apps = []models.AppParams{
					models.AppParams{
						Name:        &appName,
						DockerImage: &dockerImage,
						Path:        &path,
					},
				}
			})
			It("displays error", func() {
				errs := actor.ValidateAppParams(apps)
				Expect(errs).To(HaveLen(1))
				Expect(errs[0].Error()).To(Equal("Application my-app must not be configured with both 'docker' and 'path'"))
			})
		})

		Context("when 'routes' is provided", func() {
			BeforeEach(func() {
				appName := "my-app"
				apps = []models.AppParams{
					models.AppParams{
						Name: &appName,
						Routes: []models.ManifestRoute{
							models.ManifestRoute{
								Route: "route-name.example.com",
							},
							models.ManifestRoute{
								Route: "other-route-name.example.com",
							},
						},
					},
				}
			})

			Context("and 'hosts' is provided", func() {
				BeforeEach(func() {
					apps[0].Hosts = []string{"host-name"}
				})

				It("returns an error", func() {
					errs := actor.ValidateAppParams(apps)
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Error()).To(Equal("Application my-app must not be configured with both 'routes' and 'host'/'hosts'"))
				})
			})

			Context("and 'domains' is provided", func() {
				BeforeEach(func() {
					apps[0].Domains = []string{"domain-name"}
				})

				It("returns an error", func() {
					errs := actor.ValidateAppParams(apps)
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Error()).To(Equal("Application my-app must not be configured with both 'routes' and 'domain'/'domains'"))
				})
			})

			Context("and 'no-hostname' is provided", func() {
				BeforeEach(func() {
					noHostBool := true
					apps[0].NoHostname = &noHostBool
				})

				It("returns an error", func() {
					errs := actor.ValidateAppParams(apps)
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Error()).To(Equal("Application my-app must not be configured with both 'routes' and 'no-hostname'"))
				})
			})

			Context("and 'no-hostname' is not provided", func() {
				BeforeEach(func() {
					apps[0].NoHostname = nil
				})

				It("returns an error", func() {
					errs := actor.ValidateAppParams(apps)
					Expect(errs).To(HaveLen(0))
				})
			})
		})
	})

	Describe("MapManifestRoute", func() {
		It("passes arguments to route actor", func() {
			appName := "app-name"
			app := models.Application{
				ApplicationFields: models.ApplicationFields{
					Name: appName,
					GUID: "app-guid",
				},
			}
			appParamsFromContext := models.AppParams{
				Name: &appName,
			}

			_ = actor.MapManifestRoute("route-name.example.com/testPath", app, appParamsFromContext)
			actualRoute, actualApp, actualAppParams := routeActor.FindAndBindRouteArgsForCall(0)
			Expect(actualRoute).To(Equal("route-name.example.com/testPath"))
			Expect(actualApp).To(Equal(app))
			Expect(actualAppParams).To(Equal(appParamsFromContext))
		})
	})
})
