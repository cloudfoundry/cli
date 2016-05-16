package api

import (
	"archive/zip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"mime/multipart"
	gonet "net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/appfiles"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

//go:generate counterfeiter . BuildpackBitsRepository

type BuildpackBitsRepository interface {
	UploadBuildpack(buildpack models.Buildpack, dir string) (apiErr error)
}

type CloudControllerBuildpackBitsRepository struct {
	config       coreconfig.Reader
	gateway      net.Gateway
	zipper       appfiles.Zipper
	TrustedCerts []tls.Certificate
}

func NewCloudControllerBuildpackBitsRepository(config coreconfig.Reader, gateway net.Gateway, zipper appfiles.Zipper) (repo CloudControllerBuildpackBitsRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.zipper = zipper
	return
}

func (repo CloudControllerBuildpackBitsRepository) UploadBuildpack(buildpack models.Buildpack, buildpackLocation string) (apiErr error) {
	fileutils.TempFile("buildpack-upload", func(zipFileToUpload *os.File, err error) {
		if err != nil {
			apiErr = fmt.Errorf("%s: %s", T("Couldn't create temp file for upload"), err.Error())
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
				apiErr = fmt.Errorf("%s: %s", T("Error opening buildpack file"), statError.Error())
				err = statError
				return
			}

			if stats.IsDir() {
				buildpackFileName += ".zip" // FIXME: remove once #71167394 is fixed
				err = repo.zipper.Zip(buildpackLocation, zipFileToUpload)
			} else {
				specifiedFile, openError := os.Open(buildpackLocation)
				if openError != nil {
					apiErr = fmt.Errorf("%s: %s", T("Couldn't open buildpack file"), openError.Error())
					err = openError
					return
				}
				err = normalizeBuildpackArchive(specifiedFile, zipFileToUpload)
			}
		}

		if err != nil {
			apiErr = fmt.Errorf("%s: %s", T("Couldn't write zip file"), err.Error())
			return
		}

		apiErr = repo.uploadBits(buildpack, zipFileToUpload, buildpackFileName)
	})

	return
}

func normalizeBuildpackArchive(inputFile *os.File, outputFile *os.File) error {
	stats, toplevelErr := inputFile.Stat()
	if toplevelErr != nil {
		return toplevelErr
	}

	reader, toplevelErr := zip.NewReader(inputFile, stats.Size())
	if toplevelErr != nil {
		return toplevelErr
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

			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}

			err = r.Close()
			if err != nil {
				return err
			}
		}
	}

	toplevelErr = writer.Close()
	if toplevelErr != nil {
		return toplevelErr
	}

	_, toplevelErr = outputFile.Seek(0, 0)
	if toplevelErr != nil {
		return toplevelErr
	}

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
				Dial:            (&gonet.Dialer{Timeout: 5 * time.Second}).Dial,
				TLSClientConfig: &tls.Config{RootCAs: certPool},
				Proxy:           http.ProxyFromEnvironment,
			},
		}

		response, err := client.Get(url)
		if err != nil {
			cb(nil, err)
			return
		}
		defer response.Body.Close()

		_, err = io.Copy(tempfile, response.Body)
		if err != nil {
			cb(nil, err)
			return
		}

		_, err = tempfile.Seek(0, 0)
		if err != nil {
			cb(nil, err)
			return
		}

		cb(tempfile, nil)
	})
}

func (repo CloudControllerBuildpackBitsRepository) uploadBits(buildpack models.Buildpack, body io.Reader, buildpackName string) error {
	return repo.performMultiPartUpload(
		fmt.Sprintf("%s/v2/buildpacks/%s/bits", repo.config.APIEndpoint(), buildpack.GUID),
		"buildpack",
		buildpackName,
		body)
}

func (repo CloudControllerBuildpackBitsRepository) performMultiPartUpload(url string, fieldName string, fileName string, body io.Reader) error {
	var capturedErr error

	fileutils.TempFile("requests", func(requestFile *os.File, err error) {
		if err != nil {
			capturedErr = err
			return
		}

		writer := multipart.NewWriter(requestFile)
		part, err := writer.CreateFormFile(fieldName, fileName)

		if err != nil {
			_ = writer.Close()
			capturedErr = err
			return
		}

		_, err = io.Copy(part, body)
		if err != nil {
			capturedErr = fmt.Errorf("%s: %s", T("Error creating upload"), err.Error())
			return
		}

		err = writer.Close()
		if err != nil {
			capturedErr = err
			return
		}

		var request *net.Request
		request, err = repo.gateway.NewRequestForFile("PUT", url, repo.config.AccessToken(), requestFile)
		if err != nil {
			capturedErr = err
			return
		}

		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary())
		request.HTTPReq.Header.Set("Content-Type", contentType)

		_, err = repo.gateway.PerformRequest(request)
		if err != nil {
			capturedErr = err
		}
	})

	return capturedErr
}
