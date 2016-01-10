package actors_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/actors"
	fakeBits "github.com/cloudfoundry/cli/cf/api/application_bits/fakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/app_files/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Push Actor", func() {
	var (
		appBitsRepo  *fakeBits.FakeApplicationBitsRepository
		appFiles     *fakes.FakeAppFiles
		fakezipper   *fakes.FakeZipper
		actor        actors.PushActor
		fixturesDir  string
		appDir       string
		allFiles     []models.AppFileFields
		presentFiles []resources.AppFileResource
	)

	BeforeEach(func() {
		appBitsRepo = &fakeBits.FakeApplicationBitsRepository{}
		appFiles = &fakes.FakeAppFiles{}
		fakezipper = &fakes.FakeZipper{}
		actor = actors.NewPushActor(appBitsRepo, fakezipper, appFiles)
		fixturesDir = filepath.Join("..", "..", "fixtures", "applications")
		allFiles = []models.AppFileFields{
			models.AppFileFields{Path: "example-app/.cfignore"},
			models.AppFileFields{Path: "example-app/app.rb"},
			models.AppFileFields{Path: "example-app/config.ru"},
			models.AppFileFields{Path: "example-app/Gemfile"},
			models.AppFileFields{Path: "example-app/Gemfile.lock"},
			models.AppFileFields{Path: "example-app/ignore-me"},
			models.AppFileFields{Path: "example-app/manifest.yml"},
		}
	})

	Describe("GatherFiles", func() {
		var tmpDir string

		BeforeEach(func() {
			presentFiles = []resources.AppFileResource{
				resources.AppFileResource{Path: "example-app/ignore-me"},
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
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir)
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
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir)
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
				_, _, err := actor.GatherFiles(allFiles, appDir, tmpDir)
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

			actualFiles, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
			Expect(err).NotTo(HaveOccurred())

			expectedFiles := []resources.AppFileResource{
				resources.AppFileResource{
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

			actualFiles, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
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
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
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
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, appDir, uploadDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(Equal(expectedFiles))
				Expect(appDir).To(Equal(fixturesDir))
				Expect(uploadDir).To(Equal(tmpDir))
			})
		})

		Context("when there are local files that aren't matched", func() {
			var remoteFiles []resources.AppFileResource

			BeforeEach(func() {
				remoteFiles = []resources.AppFileResource{
					resources.AppFileResource{Path: "example-app/manifest.yml"},
				}

				appBitsRepo.GetApplicationFilesReturns(remoteFiles, nil)
			})

			It("returns true for hasFileToUpload", func() {
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
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
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, appDir, uploadDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(Equal(expectedFiles))
				Expect(appDir).To(Equal(fixturesDir))
				Expect(uploadDir).To(Equal(tmpDir))
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
				_, hasFileToUpload, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(hasFileToUpload).To(BeFalse())
			})

			It("copies nothing to the upload dir", func() {
				_, _, err := actor.GatherFiles(allFiles, fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(appFiles.CopyFilesCallCount()).To(Equal(1))
				filesToUpload, appDir, uploadDir := appFiles.CopyFilesArgsForCall(0)
				Expect(filesToUpload).To(BeEmpty())
				Expect(appDir).To(Equal(fixturesDir))
				Expect(uploadDir).To(Equal(tmpDir))
			})
		})
	})

	Describe(".UploadApp", func() {
		It("Simply delegates to the UploadApp function on the app bits repo, which is not worth testing", func() {})
	})

	Describe("ProcessPath", func() {
		var (
			wasCalled     bool
			wasCalledWith string
		)

		BeforeEach(func() {
			zipper := &app_files.ApplicationZipper{}
			actor = actors.NewPushActor(appBitsRepo, zipper, appFiles)
		})

		Context("when given a zip file", func() {
			var zipFile string

			BeforeEach(func() {
				zipFile = filepath.Join(fixturesDir, "example-app.zip")
			})

			It("extracts the zip when given a zip file", func() {
				f := func(tempDir string) {
					for _, file := range allFiles {
						actualFilePath := filepath.Join(tempDir, file.Path)
						_, err := os.Stat(actualFilePath)
						Expect(err).NotTo(HaveOccurred())
					}
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls the provided function with the directory that it extracted to", func() {
				f := func(tempDir string) {
					wasCalled = true
					wasCalledWith = tempDir
				}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).NotTo(HaveOccurred())
				Expect(wasCalled).To(BeTrue())
				Expect(wasCalledWith).NotTo(Equal(zipFile))
			})

			It("cleans up the directory that it extracted to", func() {
				var tempDirWas string
				f := func(tempDir string) {
					tempDirWas = tempDir
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
				actor = actors.NewPushActor(appBitsRepo, fakezipper, appFiles)

				f := func(tempDir string) {}
				err := actor.ProcessPath(zipFile, f)
				Expect(err).To(HaveOccurred())
			})
		})

		It("calls the provided function with the provided directory", func() {
			appDir = filepath.Join(fixturesDir, "example-app")
			f := func(tempDir string) {
				wasCalled = true
				wasCalledWith = tempDir
			}
			err := actor.ProcessPath(appDir, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalled).To(BeTrue())
			Expect(wasCalledWith).To(Equal(appDir))
		})

		It("dereferences the symlink when given a symlink to an app dir", func() {
			if runtime.GOOS == "windows" {
				Skip("This should not run on Windows")
			}

			symlink := filepath.Join(fixturesDir, "example-app-symlink")
			expectedDir := filepath.Join(fixturesDir, "example-app") // example-app-symlink -> example-app
			f := func(dir string) {
				wasCalled = true
				wasCalledWith = dir
			}

			err := actor.ProcessPath(symlink, f)
			Expect(err).NotTo(HaveOccurred())
			Expect(wasCalled).To(BeTrue())
			Expect(wasCalledWith).To(Equal(expectedDir))
		})
	})
})
