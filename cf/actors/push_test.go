package actors_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/actors"
	fakeBits "github.com/cloudfoundry/cli/cf/api/application_bits/fakes"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/app_files/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
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
		BeforeEach(func() {
			presentFiles = []resources.AppFileResource{
				resources.AppFileResource{Path: "example-app/ignore-me"},
			}

			appDir = filepath.Join(fixturesDir, "example-app.zip")
			appFiles.AppFilesInDirReturns(allFiles, nil)
			appBitsRepo.GetApplicationFilesReturns(presentFiles, nil)
		})

		It("returns files to upload with file mode unchanged", func() {
			if runtime.GOOS == "windows" {
				Skip("This does not run on windows")
			}

			info, err := os.Lstat(filepath.Join(fixturesDir, "example-app/ignore-me"))
			Expect(err).NotTo(HaveOccurred())

			expectedFileMode := fmt.Sprintf("%#o", info.Mode())

			fileutils.TempDir("gather-files", func(tmpDir string, err error) {
				actualFiles, _, err := actor.GatherFiles(fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())

				expectedFiles := []resources.AppFileResource{
					resources.AppFileResource{
						Path: "example-app/ignore-me",
						Mode: expectedFileMode,
					},
				}

				Expect(actualFiles).To(Equal(expectedFiles))
			})
		})

		It("returns files to upload with file mode always being executable", func() {
			if runtime.GOOS != "windows" {
				Skip("This runs only on windows")
			}

			info, err := os.Lstat(filepath.Join(fixturesDir, "example-app/ignore-me"))
			Expect(err).NotTo(HaveOccurred())

			expectedFileMode := fmt.Sprintf("%#o", info.Mode()|0700)

			fileutils.TempDir("gather-files", func(tmpDir string, err error) {
				actualFiles, _, err := actor.GatherFiles(fixturesDir, tmpDir)
				Expect(err).NotTo(HaveOccurred())

				expectedFiles := []resources.AppFileResource{
					resources.AppFileResource{
						Path: "example-app/ignore-me",
						Mode: expectedFileMode,
					},
				}

				Expect(actualFiles).To(Equal(expectedFiles))
			})
		})

		It("returns an error if it cannot walk the files", func() {
			fileutils.TempDir("gather-files", func(tmpDir string, err error) {
				appFiles.AppFilesInDirReturns(nil, errors.New("error"))
				_, _, err = actor.GatherFiles(appDir, tmpDir)
				Expect(err).To(HaveOccurred())
			})
		})

		It("returns an error if we cannot reach the cc", func() {
			fileutils.TempDir("gather-files", func(tmpDir string, err error) {
				appBitsRepo.GetApplicationFilesReturns(nil, errors.New("error"))
				_, _, err = actor.GatherFiles(appDir, tmpDir)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when using .cfignore", func() {
			BeforeEach(func() {
				appBitsRepo.GetApplicationFilesReturns(nil, nil)
				appDir = filepath.Join(fixturesDir, "exclude-a-default-cfignore")
			})

			It("includes the .cfignore file in the upload directory", func() {
				fileutils.TempDir("gather-files", func(tmpDir string, err error) {
					files, _, err := actor.GatherFiles(appDir, tmpDir)
					Expect(err).NotTo(HaveOccurred())

					_, err = os.Stat(filepath.Join(tmpDir, ".cfignore"))
					Expect(os.IsNotExist(err)).To(BeFalse())
					Expect(len(files)).To(Equal(0))
				})
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
	})
})
