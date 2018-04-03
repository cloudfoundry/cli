package tarball

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	biui "github.com/cloudfoundry/bosh-cli/ui"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Source interface {
	GetURL() string
	GetSHA1() string
	Description() string
}

type Provider interface {
	Get(Source, biui.Stage) (path string, err error)
}

type provider struct {
	cache            Cache
	fs               boshsys.FileSystem
	httpClient       *httpclient.HTTPClient
	downloadAttempts int
	delayTimeout     time.Duration
	logger           boshlog.Logger
	logTag           string
}

func NewProvider(
	cache Cache,
	fs boshsys.FileSystem,
	httpClient *httpclient.HTTPClient,
	downloadAttempts int,
	delayTimeout time.Duration,
	logger boshlog.Logger,
) Provider {
	return &provider{
		cache:            cache,
		fs:               fs,
		httpClient:       httpClient,
		downloadAttempts: downloadAttempts,
		delayTimeout:     delayTimeout,

		logTag: "tarballProvider",
		logger: logger,
	}
}

func (p *provider) Get(source Source, stage biui.Stage) (string, error) {
	u, err := url.Parse(source.GetURL())
	if err != nil {
		return "", bosherr.WrapError(err, "URL could not be parsed")
	}

	if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "file" && u.Scheme != "" {
		return "", bosherr.Errorf("Unsupported scheme in URL '%s'", source.GetURL())
	}

	if strings.HasPrefix(source.GetURL(), "http") {
		err := stage.Perform(fmt.Sprintf("Downloading %s", source.Description()), func() error {
			cachedPath, found := p.cache.Get(source)
			if found {
				p.logger.Debug(p.logTag, "Using the tarball from cache: '%s'", cachedPath)
				return biui.NewSkipStageError(bosherr.Error("Already downloaded"), "Found in local cache")
			}

			retryStrategy := boshretry.NewAttemptRetryStrategy(
				p.downloadAttempts, p.delayTimeout, p.downloadRetryable(source), p.logger)

			err := retryStrategy.Try()
			if err != nil {
				return bosherr.WrapErrorf(err, "Failed to download from '%s'", source.GetURL())
			}

			p.logger.Debug(p.logTag, "Using the downloaded tarball: '%s'", cachedPath)

			return nil
		})
		if err != nil {
			return "", err
		}

		return p.cache.Path(source), nil
	}

	filePath := strings.TrimPrefix(source.GetURL(), "file://")

	expandedPath, err := p.fs.ExpandPath(filePath)
	if err != nil {
		p.logger.Warn(p.logTag, "Failed to expand file path %s, using original URL", filePath)
		return filePath, nil
	}

	p.logger.Debug(p.logTag, "Using the tarball from file source: '%s'", filePath)

	return expandedPath, nil
}

func (p *provider) downloadRetryable(source Source) boshretry.Retryable {
	return boshretry.NewRetryable(func() (bool, error) {
		downloadedFile, err := p.fs.TempFile("tarballProvider")
		if err != nil {
			return true, bosherr.WrapError(err, "Unable to create temporary file")
		}

		defer func() {
			downloadedFile.Close()

			if err = p.fs.RemoveAll(downloadedFile.Name()); err != nil {
				p.logger.Warn(p.logTag, "Failed to remove downloaded file: %s", err.Error())
			}
		}()

		response, err := p.httpClient.Get(source.GetURL())
		if err != nil {
			return true, bosherr.WrapError(err, "Unable to download")
		}

		defer func() {
			if err = response.Body.Close(); err != nil {
				p.logger.Warn(p.logTag, "Failed to close download response body: %s", err.Error())
			}
		}()

		_, err = io.Copy(downloadedFile, response.Body)
		if err != nil {
			return true, bosherr.WrapError(err, "Saving downloaded bits to temporary file")
		}

		digest, err := boshcrypto.ParseMultipleDigest(source.GetSHA1())
		if err != nil {
			return true, err
		}

		err = digest.VerifyFilePath(downloadedFile.Name(), p.fs)
		if err != nil {
			return true, bosherr.WrapError(err, "Verifying digest for downloaded file")
		}

		downloadedFile.Close()

		err = p.cache.Save(downloadedFile.Name(), source)
		if err != nil {
			return true, bosherr.WrapError(err, "Saving downloaded file in cache")
		}

		return false, nil
	})
}
