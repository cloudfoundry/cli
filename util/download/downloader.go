package download

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Downloader struct {
	HTTPClient HTTPClient
	// ProgressBar ProgressBar
}

func NewDownloader(dialTimeout time.Duration) *Downloader {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   dialTimeout,
		}).DialContext,
	}

	return &Downloader{
		// ProgressBar: pb.New(0),
		HTTPClient: &http.Client{
			Transport: tr,
		},
	}
}

func (downloader Downloader) Download(url string) (string, error) {
	dir, err := ioutil.TempDir("", "buildpack-tempDir")
	if err != nil {
		return "", err
	}

	bpFileName := filepath.Join(dir, filepath.Base(url))

	resp, err := downloader.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		rawBytes, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}

		return "", RawHTTPStatusError{
			Status:      resp.Status,
			RawResponse: rawBytes,
		}
	}

	file, err := os.Create(bpFileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return bpFileName, nil
}
