package testrunner

import (
	"os/exec"
	"time"

	"github.com/tedsuo/ifrit/ginkgomon"
)

func New(binPath string, configPath string) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:              "ssh-proxy",
		AnsiColorCode:     "1;95m",
		StartCheck:        "ssh-proxy.started",
		StartCheckTimeout: 10 * time.Second,
		Command:           exec.Command(binPath, "-config="+configPath),
	})
}
