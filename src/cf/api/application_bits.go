package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type ApplicationBitsRepository interface {
	UploadApp(app cf.Application, dir string) (apiStatus net.ApiStatus)
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

func (repo CloudControllerApplicationBitsRepository) UploadApp(app cf.Application, dir string) (apiStatus net.ApiStatus) {
	dir, resourcesJson, apiStatus := repo.createUploadDir(app, dir)
	if apiStatus.NotSuccessful() {
		return
	}

	zipBuffer, err := repo.zipper.Zip(dir)
	if err != nil {
		return
	}

	apiStatus = repo.uploadBits(app, zipBuffer, resourcesJson)
	if apiStatus.NotSuccessful() {
		return
	}

	return
}

func (repo CloudControllerApplicationBitsRepository) uploadBits(app cf.Application, zipBuffer *bytes.Buffer, resourcesJson []byte) (apiStatus net.ApiStatus) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.Target, app.Guid)

	body, boundary, err := createApplicationUploadBody(zipBuffer, resourcesJson)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error creating upload", err)
		return
	}

	request, apiStatus := repo.gateway.NewRequest("PUT", url, repo.config.AccessToken, body)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	request.Header.Set("Content-Type", contentType)
	if apiStatus.NotSuccessful() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationBitsRepository) createUploadDir(app cf.Application, appDir string) (uploadDir string, resourcesJson []byte, apiStatus net.ApiStatus) {
	var err error

	// If appDir is a zip, first extract it to a temporary directory
	if fileIsZip(appDir) {
		appDir, err = extractZip(app, appDir)
		if err != nil {
			apiStatus = net.NewApiStatusWithError("Error extracting archive", err)
			return
		}
	}

	// Find which files need to be uploaded
	allAppFiles, err := cf.AppFilesInDir(appDir)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error listing app files", err)
		return
	}

	appFilesToUpload, resourcesJson, apiStatus := repo.getFilesToUpload(allAppFiles)
	if apiStatus.NotSuccessful() {
		return
	}

	// Copy files into a temporary directory and return it
	uploadDir = cf.TempDirForApp(app)

	err = cf.InitializeDir(uploadDir)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error creating upload directory", err)
		return
	}

	err = cf.CopyFiles(appFilesToUpload, appDir, uploadDir)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Error copying files to temp directory", err)
		return
	}

	return
}

func fileIsZip(file string) bool {
	isZip := strings.HasSuffix(file, ".zip")
	isWar := strings.HasSuffix(file, ".war")
	isJar := strings.HasSuffix(file, ".jar")

	return isZip || isWar || isJar
}

func extractZip(app cf.Application, zipFile string) (destDir string, err error) {
	destDir = cf.TempDirForApp(app) + "-zip"
	err = cf.InitializeDir(destDir)
	if err != nil {
		return
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return
	}
	defer r.Close()

	for _, f := range r.File {
		// Don't try to extract directories
		if f.FileInfo().IsDir() {
			continue
		}

		var (
			rc       io.ReadCloser
			destFile *os.File
		)

		rc, err = f.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		destFilePath := filepath.Join(destDir, f.Name)
		err = os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm|os.ModeDir)
		if err != nil {
			return
		}

		destFile, err = os.Create(destFilePath)
		if err != nil {
			return
		}

		_, err = io.Copy(destFile, rc)
		if err != nil {
			return
		}

		rc.Close()
	}
	return
}

func (repo CloudControllerApplicationBitsRepository) getFilesToUpload(allAppFiles []cf.AppFile) (appFilesToUpload []cf.AppFile, resourcesJson []byte, apiStatus net.ApiStatus) {
	appFilesRequest := []AppFile{}
	for _, file := range allAppFiles {
		appFilesRequest = append(appFilesRequest, AppFile{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	resourcesJson, err := json.Marshal(appFilesRequest)
	if err != nil {
		apiStatus = net.NewApiStatusWithError("Failed to create json for resource_match request", err)
		return
	}

	path := fmt.Sprintf("%s/v2/resource_match", repo.config.Target)
	req, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, bytes.NewReader(resourcesJson))
	if apiStatus.NotSuccessful() {
		return
	}

	res := []AppFile{}
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(req, &res)

	appFilesToUpload = make([]cf.AppFile, len(allAppFiles))
	copy(appFilesToUpload, allAppFiles)
	for _, file := range res {
		appFile := cf.AppFile{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		}
		appFilesToUpload = deleteAppFile(appFilesToUpload, appFile)
	}

	return
}

func deleteAppFile(appFiles []cf.AppFile, targetFile cf.AppFile) []cf.AppFile {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}

func createApplicationUploadBody(zipBuffer *bytes.Buffer, resourcesJson []byte) (body *bytes.Buffer, boundary string, err error) {
	body = new(bytes.Buffer)

	writer := multipart.NewWriter(body)
	defer writer.Close()

	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBuffer(resourcesJson))
	if err != nil {
		return
	}

	if zipBuffer.Len() == 0 {
		return
	}

	part, err = createZipPartWriter(zipBuffer, writer)
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipBuffer)
	if err != nil {
		return
	}

	return
}

func createZipPartWriter(zipBuffer *bytes.Buffer, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipBuffer.Len()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}
