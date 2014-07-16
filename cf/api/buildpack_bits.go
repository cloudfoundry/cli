package api

import (
	"archive/zip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
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
	UploadBuildpack(buildpack models.Buildpack, dir string) (apiErr error)
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

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, buildpackLocation string) (apiErr error) {
	fileutils.TempFile("buildpack-upload", func(zipFileToUpload *os.File, err error) {
		if err != nil {
			apiErr = errors.NewWithError(T("Couldn't create temp file for upload"), err)
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

			stats, statError := os.Stat(buildpackLocation)
			if statError != nil {
				apiErr = errors.NewWithError(T("Error opening buildpack file"), statError)
				err = statError
				return
			}

			if stats.IsDir() {
				buildpackFileName += ".zip" // FIXME: remove once #71167394 is fixed
				err = repo.zipper.Zip(buildpackLocation, zipFileToUpload)
			} else {
				specifiedFile, openError := os.Open(buildpackLocation)
				if openError != nil {
					apiErr = errors.NewWithError(T("Couldn't open buildpack file"), openError)
					err = openError
					return
				}
				err = normalizeBuildpackArchive(specifiedFile, zipFileToUpload)
			}
		}

		if err != nil {
			apiErr = errors.NewWithError(T("Couldn't write zip file"), err)
			return
		}

		apiErr = repo.uploadBits(buildpack, zipFileToUpload, buildpackFileName)
	})

	return
}

func normalizeBuildpackArchive(inputFile *os.File, outputFile *os.File) error {
	stats, err := inputFile.Stat()
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(inputFile, stats.Size())
	if err != nil {
		return err
	}

	contents := reader.File

	parentPath, hasBuildpack := findBuildpackPath(contents)

	if !hasBuildpack {
		return errors.New(T("Zip archive does not contain a buildpack"))
	}

	writer := zip.NewWriter(outputFile)

	for _, file := range contents {
		name := file.Name
		if strings.HasPrefix(name, parentPath) {
			relativeFilename := strings.TrimPrefix(name, parentPath+"/")
			if relativeFilename == "" {
				continue
			}

			fileInfo := file.FileInfo()
			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return err
			}
			header.Name = relativeFilename

			w, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			r, err := file.Open()
			if err != nil {
				return err
			}

			io.Copy(w, r)
			err = r.Close()
			if err != nil {
				return err
			}
		}
	}

	writer.Close()
	outputFile.Seek(0, 0)
	return nil
}

func findBuildpackPath(zipFiles []*zip.File) (parentPath string, foundBuildpack bool) {
	needle := "bin/compile"

	for _, file := range zipFiles {
		if strings.HasSuffix(file.Name, needle) {
			foundBuildpack = true
			parentPath = path.Join(file.Name, "..", "..")
			if parentPath == "." {
				parentPath = ""
			}
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

func (repo CloudControllerBuildpackBitsRepository) uploadBits(buildpack models.Buildpack, body io.Reader, buildpackName string) error {
	return repo.performMultiPartUpload(
		fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.ApiEndpoint(), buildpack.Guid),
		"buildpack",
		buildpackName,
		body)
}

func (repo CloudControllerBuildpackBitsRepository) performMultiPartUpload(url string, fieldName string, fileName string, body io.Reader) (apiErr error) {
	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			apiErr = err
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
			apiErr = errors.NewWithError(T("Error creating upload"), err)
			return
		}

		var request *net.Request
		request, apiErr = repo.gateway.NewRequest("PUT", url, repo.config.AccessToken(), requestFile)
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
		request.HttpReq.Header.Set("Content-Type", contentType)
		if apiErr != nil {
			return
		}

		_, apiErr = repo.gateway.PerformRequest(request)
	})

	return
}
