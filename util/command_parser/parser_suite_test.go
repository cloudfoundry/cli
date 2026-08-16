package command_parser_test

import (
	"io"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Parser Suite")
}

var _ = BeforeSuite(func() {
	// Suppress log output during tests. This equates with setting to panic-level severity in older logging libraries
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
})
