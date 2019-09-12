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
		HTTPClient: &http.Client{
			Transport: tr,
		},
	}
}

func (downloader Downloader) Download(url string, tmpDirPath string) (string, error) {
	bpFileName := filepath.Join(tmpDirPath, filepath.Base(url))

	resp, err := downloader.HTTPClient.Get(url)
	if err != nil {
		return bpFileName, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		rawBytes, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			return bpFileName, readErr
		}
		return bpFileName, RawHTTPStatusError{
			Status:      resp.Status,
			RawResponse: rawBytes,
		}
	}

	file, err := os.Create(bpFileName)
	if err != nil {
		return bpFileName, err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return bpFileName, err
	}

	return bpFileName, nil
}
