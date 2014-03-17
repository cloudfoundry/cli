package api

import (
	"archive/zip"
	"cf/app_files"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack models.Buildpack, dir string) (apiErr errors.Error)
}

type CloudControllerBuildpackBitsRepository struct {
	config       configuration.Reader
	gateway      net.Gateway
	zipper       app_files.Zipper
	TrustedCerts []tls.Certificate
}

func NewCloudControllerBuildpackBitsRepository(config configuration.Reader, gateway net.Gateway, zipper app_files.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, buildpackLocation string) (apiErr errors.Error) {
	fileutils.TempFile("buildpack-upload", func(zipFileToUpload *os.File, err error) {
		if err != nil {
			apiErr = errors.NewErrorWithError("Couldn't create temp file for upload", err)
			return
		}

		var buildpackFileName string
		if isWebURL(buildpackLocation) {
			buildpackFileName = path.Base(buildpackLocation)
			repo.downloadBuildpack(buildpackLocation, func(downloadFile *os.File, downloadErr error) {
				if downloadErr != nil {
					err = downloadErr
					return
				}

				err = normalizeBuildpackArchive(downloadFile, zipFileToUpload)
			})
		} else {
			buildpackFileName = filepath.Base(buildpackLocation)

			stats, err := os.Stat(buildpackLocation)
			if err != nil {
				apiErr = errors.NewErrorWithError("Error opening buildpack file", err)
				return
			}

			if stats.IsDir() {
				err = repo.zipper.Zip(buildpackLocation, zipFileToUpload)
			} else {
				specifiedFile, err := os.Open(buildpackLocation)
				if err != nil {
					apiErr = errors.NewErrorWithError("Couldn't open buildpack file", err)
					return
				}
				err = normalizeBuildpackArchive(specifiedFile, zipFileToUpload)
			}
		}

		if err != nil {
			apiErr = errors.NewErrorWithError("Couldn't write zip file", err)
			return
		}

		apiErr = repo.uploadBits(buildpack, zipFileToUpload, buildpackFileName)
	})

	return
}

func normalizeBuildpackArchive(inputFile *os.File, outputFile *os.File) (err error) {
	stats, err := inputFile.Stat()
	if err != nil {
		return
	}

	reader, err := zip.NewReader(inputFile, stats.Size())
	if err != nil {
		return
	}

	contents := reader.File

	parentPath, hasBuildpack := findBuildpackPath(contents)

	if !hasBuildpack {
		return errors.New("Zip archive does not contain a buildpack")
	}

	writer := zip.NewWriter(outputFile)

	for _, file := range contents {
		name := file.Name
		if parentPath == "." || strings.HasPrefix(name, parentPath) {
			var (
				r      io.ReadCloser
				w      io.Writer
				header *zip.FileHeader
			)

			fileInfo := file.FileInfo()
			header, err = zip.FileInfoHeader(fileInfo)
			header.Name = strings.Replace(name, parentPath+"/", "", 1)

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
	needle := "bin/compile"

	for _, file := range zipFiles {
		if strings.HasSuffix(file.Name, needle) {
			foundBuildpack = true
			parentPath = path.Join(file.Name, "..", "..")
			return
		}
	}
	return
}

func isWebURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func (repo CloudControllerBuildpackBitsRepository) downloadBuildpack(url string, cb func(*os.File, error)) {
	fileutils.TempFile("buildpack-download", func(tempfile *os.File, err error) {
		if err != nil {
			cb(nil, err)
			return
		}

		var certPool *x509.CertPool
		if len(repo.TrustedCerts) > 0 {
			certPool = x509.NewCertPool()
			for _, tlsCert := range repo.TrustedCerts {
				cert, _ := x509.ParseCertificate(tlsCert.Certificate[0])
				certPool.AddCert(cert)
			}
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: certPool},
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

func (repo CloudControllerBuildpackBitsRepository) uploadBits(buildpack models.Buildpack, body io.Reader, buildpackName string) errors.Error {
	return repo.performMultiPartUpload(
		fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.ApiEndpoint(), buildpack.Guid),
		"buildpack",
		buildpackName,
		body)
}

func (repo CloudControllerBuildpackBitsRepository) performMultiPartUpload(url string, fieldName string, fileName string, body io.Reader) (apiErr errors.Error) {
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiErr = errors.NewErrorWithMessage(err.Error())
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
			apiErr = errors.NewErrorWithError("Error creating upload", err)
			return
		}

		var request *net.Request
		request, apiErr = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
		request.HttpReq.Header.Set("Content-Type", contentType)
		if apiErr != nil {
			return
		}

		apiErr = repo.gateway.PerformRequest(request)
	})

	return
}
