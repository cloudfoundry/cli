package api

import (
	"archive/zip"
	"bytes"
	"cf/app_files"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"encoding/json"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"time"
)

type AppFileResource struct {
	Path string `json:"fn"`
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}

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

func (repo CloudControllerApplicationBitsRepository) UploadApp(appGuid string, appDir string, cb func(path string, zipSize, fileCount uint64)) (apiErr error) {
	fileutils.TempDir("apps", func(uploadDir string, err error) {
		if err != nil {
			apiErr = err
			return
		}

		var presentResourcesJson []byte
		repo.sourceDir(appDir, func(sourceDir string, sourceErr error) {
			if sourceErr != nil {
				err = sourceErr
				return
			}
			presentResourcesJson, err = repo.copyUploadableFiles(sourceDir, uploadDir)
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

			err = repo.zipper.Zip(uploadDir, zipFile)
			if err != nil {
				apiErr = errors.NewWithError("Error zipping application", err)
				return
			}

			stat, err := zipFile.Stat()
			if err != nil {
				apiErr = errors.NewWithError("Error zipping application", err)
				return
			}
			cb(appDir, uint64(stat.Size()), app_files.CountFiles(uploadDir))

			apiErr = repo.uploadBits(appGuid, zipFile, presentResourcesJson)
			if apiErr != nil {
				return
			}
		})
	})
	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(appGuid string, zipFile *os.File, presentResourcesJson []byte) (apiErr error) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.ApiEndpoint(), appGuid)
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiErr = errors.NewWithError("Error creating tmp file: %s", err)
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile, presentResourcesJson)
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

		response := &Resource{}
		_, apiErr = repo.gateway.PerformPollingRequestForJSONResponse(request, response, 5*time.Minute)
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

func (repo CloudControllerApplicationBitsRepository) copyUploadableFiles(appDir string, uploadDir string) (presentResourcesJson []byte, err error) {
	// Find which files need to be uploaded
	allAppFiles, err := app_files.AppFilesInDir(appDir)
	if err != nil {
		return
	}

	appFilesToUpload, presentResourcesJson, apiErr := repo.getFilesToUpload(allAppFiles)
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

			err = fileutils.SetExecutableBits(destFilePath, f.FileInfo())
			if err != nil {
				return
			}
		}()
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) getFilesToUpload(allAppFiles []models.AppFileFields) (appFilesToUpload []models.AppFileFields, presentResourcesJson []byte, apiErr error) {
	appFilesRequest := []AppFileResource{}
	for _, file := range allAppFiles {
		appFilesRequest = append(appFilesRequest, AppFileResource{
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

	path := fmt.Sprintf("%s/v2/resource_match", repo.config.ApiEndpoint())
	req, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken(), bytes.NewReader(allAppFilesJson))
	if apiErr != nil {
		return
	}

	presentResourcesJson, _, _, apiErr = repo.gateway.PerformRequestForResponseBytes(req)

	fileResource := []AppFileResource{}
	err = json.Unmarshal(presentResourcesJson, &fileResource)

	if err != nil {
		apiErr = errors.NewWithError("Failed to unmarshal json response from resource_match request", err)
		return
	}

	appFilesToUpload = make([]models.AppFileFields, len(allAppFiles))
	copy(appFilesToUpload, allAppFiles)
	for _, file := range fileResource {
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

	zipStats, err := zipFile.Stat()
	if err != nil {
		return
	}

	if zipStats.Size() == 0 {
		return
	}

	part, err = createZipPartWriter(zipStats, writer)
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipFile)
	if err != nil {
		return
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
