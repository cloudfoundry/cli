package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type AppFileResource struct {
	Path string `json:"fn"`
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}

type ApplicationBitsRepository interface {
	UploadApp(appGuid, dir string, cb func(zipSize, fileCount uint64)) (apiResponse net.ApiResponse)
}

type CloudControllerApplicationBitsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerApplicationBitsRepository(config *configuration.Configuration, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerApplicationBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerApplicationBitsRepository) UploadApp(appGuid string, appDir string, cb func(zipSize, fileCount uint64)) (apiResponse net.ApiResponse) {
	fileutils.TempDir("apps", func(uploadDir string, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		repo.sourceDir(appDir, func(sourceDir string, sourceErr error) {
			if sourceErr != nil {
				err = sourceErr
				return
			}
			err = repo.copyUploadableFiles(sourceDir, uploadDir)
		})

		if err != nil {
			apiResponse = net.NewApiResponseWithMessage("%s", err)
			return
		}

		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			if err != nil {
				apiResponse = net.NewApiResponseWithMessage("%s", err.Error())
				return
			}

			err = repo.zipper.Zip(uploadDir, zipFile)
			if err != nil {
				apiResponse = net.NewApiResponseWithError("Error zipping application", err)
				return
			}

			stat, err := zipFile.Stat()
			if err != nil {
				apiResponse = net.NewApiResponseWithError("Error zipping application", err)
				return
			}
			cb(uint64(stat.Size()), cf.CountFiles(uploadDir))

			apiResponse = repo.uploadBits(appGuid, zipFile)
			if apiResponse.IsNotSuccessful() {
				return
			}
		})
	})
	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(appGuid string, zipFile *os.File) (apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.Target, appGuid)

	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error creating tmp file: %s", err)
			return
		}

		boundary, err := repo.writeUploadBody(zipFile, requestFile)
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error writing to tmp file: %s", err)
			return
		}

		var request *net.Request
		request, apiResponse = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, requestFile)
		if apiResponse.IsNotSuccessful() {
			return
		}

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		request.HttpReq.Header.Set("Content-Type", contentType)

		response := &Resource{}
		_, apiResponse = repo.gateway.PerformPollingRequestForJSONResponse(request, response)
		if apiResponse.IsNotSuccessful() {
			return
		}
	})

	return
}

func (repo CloudControllerApplicationBitsRepository) sourceDir(appDir string, cb func(sourceDir string, err error)) {
	// If appDir is a zip, first extract it to a temporary directory
	if !repo.fileIsZip(appDir) {
		cb(appDir, nil)
		return
	}

	fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
		if err != nil {
			cb("", err)
			return
		}

		err = repo.extractZip(appDir, tmpDir)
		cb(tmpDir, err)
	})
}

func (repo CloudControllerApplicationBitsRepository) copyUploadableFiles(appDir string, uploadDir string) (err error) {
	// Find which files need to be uploaded
	allAppFiles, err := cf.AppFilesInDir(appDir)
	if err != nil {
		return
	}

	// Copy files into a temporary directory and return it
	err = cf.CopyFiles(allAppFiles, appDir, uploadDir)
	if err != nil {
		return
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) fileIsZip(file string) bool {
	isZip := strings.HasSuffix(file, ".zip")
	isWar := strings.HasSuffix(file, ".war")
	isJar := strings.HasSuffix(file, ".jar")

	return isZip || isWar || isJar
}

func (repo CloudControllerApplicationBitsRepository) extractZip(zipFile string, destDir string) (err error) {
	r, err := zip.OpenReader(zipFile)
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

			if err != nil {
				return
			}

			var rc io.ReadCloser
			rc, err = f.Open()
			if err != nil {
				return
			}

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

func (repo CloudControllerApplicationBitsRepository) deleteAppFile(appFiles []cf.AppFileFields, targetFile cf.AppFileFields) []cf.AppFileFields {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func (repo CloudControllerApplicationBitsRepository) writeUploadBody(zipFile *os.File, body *os.File) (boundary string, err error) {
	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBufferString("[]"))
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
