package api

import (
	"archive/zip"
	"cf"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"crypto/tls"
	"errors"
	"fileutils"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack models.Buildpack, dir string) (apiResponse net.ApiResponse)
}

type CloudControllerBuildpackBitsRepository struct {
	config  configuration.Reader
	gateway net.Gateway
	zipper  cf.Zipper
}

func NewCloudControllerBuildpackBitsRepository(config configuration.Reader, gateway net.Gateway, zipper cf.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, buildpackLocation string) (apiResponse net.ApiResponse) {
	fileutils.TempFile("buildpack-upload", func(zipFileToUpload *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithError("Couldn't create temp file for upload", err)
			return
		}

		var buildpackFileName string
		if isWebURL(buildpackLocation) {
			buildpackFileName = path.Base(buildpackLocation)
			downloadBuildpack(buildpackLocation, func(downloadFile *os.File, downloadErr error) {
				if downloadErr != nil {
					err = downloadErr
					return
				}

				var stats os.FileInfo
				stats, err = downloadFile.Stat()
				if err != nil {
					return
				}

				err = normalizeBuildpackArchive(downloadFile, stats.Size(), zipFileToUpload)
			})
		} else {
			buildpackFileName = filepath.Base(buildpackLocation)

			stats, err := os.Stat(buildpackLocation)
			if err != nil {
				apiResponse = net.NewApiResponseWithError("Error opening buildpack file", err)
				return
			}

			if stats.IsDir() {
				err = repo.zipper.Zip(buildpackLocation, zipFileToUpload)
			} else {
				specifiedFile, err := os.Open(buildpackLocation)
				if err != nil {
					apiResponse = net.NewApiResponseWithError("Couldn't open buildpack file", err)
					return
				}
				err = normalizeBuildpackArchive(specifiedFile, stats.Size(), zipFileToUpload)
			}
		}

		if err != nil {
			apiResponse = net.NewApiResponseWithError("Couldn't write zip file", err)
			return
		}

		apiResponse = repo.uploadBits(buildpack, zipFileToUpload, buildpackFileName)
	})

	return
}

func normalizeBuildpackArchive(inputFile *os.File, size int64, outputFile *os.File) (err error) {
	reader, _ := zip.NewReader(inputFile, size)
	contents := reader.File

	parentPath, hasBuildpack := findBuildpackPath(contents)

	if !hasBuildpack {
		return errors.New("Zip archive does not contain a buildpack")
	}

	writer := zip.NewWriter(outputFile)

	for _, file := range contents {
		name := file.Name
		if strings.HasPrefix(name, parentPath) {
			var (
				r      io.ReadCloser
				w      io.Writer
				header *zip.FileHeader
			)

			fileInfo := file.FileInfo()
			header, err = zip.FileInfoHeader(fileInfo)
			header.Name = filepath.ToSlash(strings.Replace(name, parentPath, "", 1))

			r, err = file.Open()
			if err != nil {
				return
			}

			w, err = writer.CreateHeader(header)
			if err != nil {
				return
			}

			io.Copy(w, r)
			err = r.Close()
			if err != nil {
				return
			}
		}
	}

	writer.Close()
	outputFile.Seek(0, 0)
	return
}

func findBuildpackPath(zipFiles []*zip.File) (parentPath string, foundBuildpack bool) {
	needle := filepath.Join("bin", "compile")

	for _, file := range zipFiles {
		if strings.HasSuffix(file.Name, needle) {
			foundBuildpack = true
			parentPath = path.Join(file.Name, "..", "..") + "/"
			return
		}
	}
	return
}

func isWebURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func downloadBuildpack(url string, cb func(*os.File, error)) {
	fileutils.TempFile("buildpack-download", func(tempfile *os.File, err error) {
		if err != nil {
			cb(nil, err)
			return
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy:           http.ProxyFromEnvironment,
			},
		}
		response, err := client.Get(url)
		if err != nil {
			cb(nil, err)
			return
		}

		io.Copy(tempfile, response.Body)
		tempfile.Seek(0, 0)
		cb(tempfile, nil)
	})
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(buildpack models.Buildpack, body io.Reader, buildpackName string) net.ApiResponse {
	return repo.performMultiPartUpload(
		fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.ApiEndpoint(), buildpack.Guid),
		"buildpack",
		buildpackName,
		body)
}

func (repo CloudControllerBuildpackBitsRepository) performMultiPartUpload(url string, fieldName string, fileName string, body io.Reader) (apiResponse net.ApiResponse) {
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiResponse = net.NewApiResponseWithMessage(err.Error())
			return
		}

		writer := multipart.NewWriter(requestFile)
		part, err := writer.CreateFormFile(fieldName, fileName)

		if err != nil {
			writer.Close()
			return
		}

		_, err = io.Copy(part, body)
		writer.Close()

		if err != nil {
			apiResponse = net.NewApiResponseWithError("Error creating upload", err)
			return
		}

		var request *net.Request
		request, apiResponse = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
		request.HttpReq.Header.Set("Content-Type", contentType)
		if apiResponse.IsNotSuccessful() {
			return
		}

		apiResponse = repo.gateway.PerformRequest(request)
	})

	return
}
