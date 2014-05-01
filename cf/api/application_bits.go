package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultAppUploadBitsTimeout = 15 * time.Minute
)

type ApplicationBitsRepository interface {
	UploadApp(appGuid, dir string, cb func(path string, zipSize, fileCount uint64)) (apiErr error)
}

type CloudControllerApplicationBitsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
	zipper  app_files.Zipper
}

func NewCloudControllerApplicationBitsRepository(config configuration.Reader, gateway net.Gateway, zipper app_files.Zipper) (repo CloudControllerApplicationBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerApplicationBitsRepository) UploadApp(appGuid string, appDir string, fileSizePrinter func(path string, zipSize, fileCount uint64)) (apiErr error) {
	fileutils.TempDir("apps", func(uploadDir string, err error) {
		if err != nil {
			apiErr = err
			return
		}

		var presentFiles []resources.AppFileResource
		repo.sourceDir(appDir, func(sourceDir string, sourceErr error) {
			if sourceErr != nil {
				err = sourceErr
				return
			}
			presentFiles, err = repo.copyUploadableFiles(sourceDir, uploadDir)
		})

		if err != nil {
			apiErr = err
			return
		}

		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			if err != nil {
				apiErr = err
				return
			}

			zipFileSize := uint64(0)
			zipFileCount := uint64(0)

			err = repo.zipper.Zip(uploadDir, zipFile)
			switch err := err.(type) {
			case nil:
				stat, err := zipFile.Stat()
				if err != nil {
					apiErr = errors.NewWithError("Error zipping application", err)
					return
				}

				zipFileSize = uint64(stat.Size())
				zipFileCount = app_files.CountFiles(uploadDir)
			case *errors.EmptyDirError:
				zipFile = nil
			default:
				apiErr = errors.NewWithError("Error zipping application", err)
				return
			}

			fileSizePrinter(appDir, zipFileSize, zipFileCount)

			apiErr = repo.uploadBits(appGuid, zipFile, presentFiles)
			if apiErr != nil {
				return
			}
		})
	})
	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) (apiErr error) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.ApiEndpoint(), appGuid)
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiErr = errors.NewWithError("Error creating tmp file: %s", err)
			return
		}

		presentFilesJSON, err := json.Marshal(presentFiles)
		if err != nil {
			apiErr = errors.NewWithError("Error marshaling JSON", err)
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile, presentFilesJSON)
		if err != nil {
			apiErr = errors.NewWithError("Error writing to tmp file: %s", err)
			return
		}

		var request *net.Request
		request, apiErr = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		if apiErr != nil {
			return
		}

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)

		response := &resources.Resource{}
		_, apiErr = repo.gateway.PerformPollingRequestForJSONResponse(request, response, DefaultAppUploadBitsTimeout)
		if apiErr != nil {
			return
		}
	})

	return
}

func (repo CloudControllerApplicationBitsRepository) sourceDir(appDir string, cb func(sourceDir string, err error)) {
	// If appDir is a zip, first extract it to a temporary directory
	if repo.zipper.IsZipFile(appDir) {
		fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
			err = repo.extractZip(appDir, tmpDir)
			cb(tmpDir, err)
		})
	} else {
		cb(appDir, nil)
	}
}

func (repo CloudControllerApplicationBitsRepository) copyUploadableFiles(appDir string, uploadDir string) (presentFiles []resources.AppFileResource, err error) {
	// Find which files need to be uploaded
	allAppFiles, err := app_files.AppFilesInDir(appDir)
	if err != nil {
		return
	}

	appFilesToUpload, presentFiles, apiErr := repo.getFilesToUpload(allAppFiles)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}

	// Copy files into a temporary directory and return it
	err = app_files.CopyFiles(appFilesToUpload, appDir, uploadDir)
	if err != nil {
		return
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) extractZip(appDir, destDir string) (err error) {
	r, err := zip.OpenReader(appDir)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			// Don't try to extract directories
			if f.FileInfo().IsDir() {
				return
			}

			var rc io.ReadCloser
			rc, err = f.Open()
			if err != nil {
				return
			}

			// functional scope from above is important
			// otherwise this only closes the last file handle
			defer rc.Close()

			destFilePath := filepath.Join(destDir, f.Name)

			err = fileutils.CopyReaderToPath(rc, destFilePath)
			if err != nil {
				return
			}

			err = os.Chmod(destFilePath, f.FileInfo().Mode())
			if err != nil {
				return
			}
		}()
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) getFilesToUpload(allAppFiles []models.AppFileFields) (appFilesToUpload []models.AppFileFields, presentFiles []resources.AppFileResource, apiErr error) {
	appFilesRequest := []resources.AppFileResource{}
	for _, file := range allAppFiles {
		appFilesRequest = append(appFilesRequest, resources.AppFileResource{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	allAppFilesJson, err := json.Marshal(appFilesRequest)
	if err != nil {
		apiErr = errors.NewWithError("Failed to create json for resource_match request", err)
		return
	}

	apiErr = repo.gateway.UpdateResourceSync(
		repo.config.ApiEndpoint()+"/v2/resource_match",
		bytes.NewReader(allAppFilesJson),
		&presentFiles)

	if apiErr != nil {
		return
	}

	appFilesToUpload = make([]models.AppFileFields, len(allAppFiles))
	copy(appFilesToUpload, allAppFiles)
	for _, file := range presentFiles {
		appFile := models.AppFileFields{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		}
		appFilesToUpload = repo.deleteAppFile(appFilesToUpload, appFile)
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) deleteAppFile(appFiles []models.AppFileFields, targetFile models.AppFileFields) []models.AppFileFields {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func (repo CloudControllerApplicationBitsRepository) writeUploadBody(zipFile *os.File, body *os.File, presentResourcesJson []byte) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBuffer(presentResourcesJson))
	if err != nil {
		return
	}

	if zipFile != nil {
		zipStats, zipErr := zipFile.Stat()
		if zipErr != nil {
			return
		}

		if zipStats.Size() == 0 {
			return
		}

		part, zipErr = createZipPartWriter(zipStats, writer)
		if zipErr != nil {
			return
		}

		_, zipErr = io.Copy(part, zipFile)
		if zipErr != nil {
			return
		}
	}

	return
}

func createZipPartWriter(zipStats os.FileInfo, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipStats.Size()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}
