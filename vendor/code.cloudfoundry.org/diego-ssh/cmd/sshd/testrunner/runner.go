package testrunner

import (
	"os/exec"
	"strconv"
	"time"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type Args struct {
	Address                     string
	HostKey                     string
	AuthorizedKey               string
	AllowedCiphers              string
	AllowedMACs                 string
	AllowedKeyExchanges         string
	AllowUnauthenticatedClients bool
	InheritDaemonEnv            bool
}

func (args Args) ArgSlice() []string {
	return []string{
		"-address=" + args.Address,
		"-hostKey=" + args.HostKey,
		"-authorizedKey=" + args.AuthorizedKey,
		"-allowedCiphers=" + args.AllowedCiphers,
		"-allowedMACs=" + args.AllowedMACs,
		"-allowedKeyExchanges=" + args.AllowedKeyExchanges,
		"-allowUnauthenticatedClients=" + strconv.FormatBool(args.AllowUnauthenticatedClients),
		"-inheritDaemonEnv=" + strconv.FormatBool(args.InheritDaemonEnv),
	}
}

func New(binPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:              "sshd",
		AnsiColorCode:     "1;96m",
		StartCheck:        "sshd.started",
		StartCheckTimeout: 10 * time.Second,
		Command:           exec.Command(binPath, args.ArgSlice()...),
	})
}
