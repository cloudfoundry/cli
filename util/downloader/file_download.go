package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Downloader interface {
	DownloadFile(string) (int64, string, error)
	RemoveFile() error
	SavePath() string
}

type downloader struct {
	saveDir    string
	filename   string
	downloaded bool
}

func NewDownloader(saveDir string) Downloader {
	return &downloader{
		saveDir:    saveDir,
		downloaded: false,
	}
}

//this func returns byte written, filename and error
func (d *downloader) DownloadFile(url string) (int64, string, error) {
	c := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path

			//some redirect return '/' as url
			if strings.Trim(r.URL.Opaque, "/") != "" {
				url = r.URL.Opaque
			}

			return nil
		},
	}

	r, err := c.Get(url)

	if err != nil {
		return 0, "", err
	}
	defer r.Body.Close()

	if r.StatusCode == 200 {
		d.filename = getFilenameFromHeader(r.Header.Get("Content-Disposition"))

		if d.filename == "" {
			d.filename = getFilenameFromURL(url)
		}

		f, err := os.Create(filepath.Join(d.saveDir, d.filename))
		if err != nil {
			return 0, "", err
		}
		defer f.Close()

		size, err := io.Copy(f, r.Body)
		if err != nil {
			return 0, "", err
		}

		d.downloaded = true
		return size, d.filename, nil

	}
	return 0, "", fmt.Errorf("Error downloading file from %s", url)
}

func (d *downloader) RemoveFile() error {
	if !d.downloaded {
		return nil
	}
	d.downloaded = false
	return os.Remove(filepath.Join(d.saveDir, d.filename))
}

func getFilenameFromHeader(h string) string {
	if h == "" {
		return ""
	}

	contents := strings.Split(h, ";")
	for _, content := range contents {
		if strings.Contains(content, "filename=") {
			content = strings.TrimSpace(content)
			name := strings.TrimLeft(content, "filename=")
			return strings.Trim(name, `"`)
		}
	}

	return ""
}

func getFilenameFromURL(url string) string {
	tmp := strings.Split(url, "/")
	token := tmp[len(tmp)-1]

	if i := strings.LastIndex(token, "?"); i != -1 {
		token = token[i+1:]
	}

	if i := strings.LastIndex(token, "&"); i != -1 {
		token = token[i+1:]
	}

	if i := strings.LastIndex(token, "="); i != -1 {
		return token[i+1:]
	}

	return token
}

func (d *downloader) SavePath() string {
	return d.saveDir
}
