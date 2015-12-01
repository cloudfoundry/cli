package downloader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDownloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Downloader Suite")
}
