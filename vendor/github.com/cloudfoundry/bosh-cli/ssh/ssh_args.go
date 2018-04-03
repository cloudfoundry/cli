package ssh

import (
	"fmt"
	"net"
	"strings"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	proxy "github.com/cloudfoundry/socks5-proxy"
)

type SSHArgs struct {
	ConnOpts ConnectionOpts
	Result   boshdir.SSHResult

	ForceTTY bool

	PrivKeyFile    boshsys.File
	KnownHostsFile boshsys.File

	socks5Proxy *proxy.Socks5Proxy
	dialer      proxy.DialFunc
}

func NewSSHArgs(connOpts ConnectionOpts, result boshdir.SSHResult, forceTTY bool, privKeyFile boshsys.File, knownHostsFile boshsys.File) SSHArgs {
	socks5Proxy := proxy.NewSocks5Proxy(proxy.NewHostKeyGetter())
	boshhttpDialer := boshhttp.SOCKS5DialFuncFromEnvironment(net.Dial, socks5Proxy)
	dialer := func(net, addr string) (net.Conn, error) {
		return boshhttpDialer(net, addr)
	}

	return SSHArgs{
		ConnOpts:       connOpts,
		Result:         result,
		ForceTTY:       forceTTY,
		PrivKeyFile:    privKeyFile,
		KnownHostsFile: knownHostsFile,
		socks5Proxy:    socks5Proxy,
		dialer:         dialer,
	}
}

func (a SSHArgs) LoginForHost(host boshdir.Host) []string {
	return []string{host.Host, "-l", host.Username}
}

func (a SSHArgs) OptsForHost(host boshdir.Host) []string {
	// Options are used for both ssh and scp
	cmdOpts := []string{}

	if a.ForceTTY {
		cmdOpts = append(cmdOpts, "-tt")
	}

	cmdOpts = append(cmdOpts, []string{
		"-o", "ServerAliveInterval=30",
		"-o", "ForwardAgent=no",
		"-o", "PasswordAuthentication=no",
		"-o", "IdentitiesOnly=yes",
		"-o", "IdentityFile=" + a.PrivKeyFile.Name(),
		"-o", "StrictHostKeyChecking=yes",
		"-o", "UserKnownHostsFile=" + a.KnownHostsFile.Name(),
	}...)

	gwUsername, gwHost, gwPrivKeyPath := a.gwOpts()

	if len(a.ConnOpts.SOCKS5Proxy) > 0 {
		proxyString := a.ConnOpts.SOCKS5Proxy
		if strings.HasPrefix(proxyString, "ssh+") {
			a.socks5Proxy.StartWithDialer(a.dialer)
			proxyString, _ = a.socks5Proxy.Addr()
		}

		proxyOpt := fmt.Sprintf(
			"ProxyCommand=nc -x %s %%h %%p",
			strings.TrimPrefix(proxyString, "socks5://"),
		)

		cmdOpts = append(cmdOpts, "-o", proxyOpt)

	} else if len(gwHost) > 0 {
		gwCmdOpts := []string{
			"-o", "ServerAliveInterval=30",
			"-o", "ForwardAgent=no",
			"-o", "ClearAllForwardings=yes",
			// Strict host key checking for a gateway is not necessary
			// since ProxyCommand is only used for forwarding TCP and
			// agent forwarding is disabled
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
		}

		if len(gwPrivKeyPath) > 0 {
			gwCmdOpts = append(
				gwCmdOpts,
				"-o", "PasswordAuthentication=no",
				"-o", "IdentitiesOnly=yes",
				"-o", "IdentityFile="+gwPrivKeyPath,
			)
		}

		// It appears that when using ssh -W, IPv6 address needs to be put in brackets
		// fixes: `Bad stdio forwarding specification 'fd7a:eeed:e696:...'`
		proxyHostPortTmpl := "%h:%p"
		if strings.Contains(host.Host, ":") {
			proxyHostPortTmpl = "[%h]:%p"
		}

		proxyOpt := fmt.Sprintf(
			// Always force TTY for gateway ssh
			"ProxyCommand=ssh -tt -W %s -l %s %s %s",
			proxyHostPortTmpl,
			gwUsername,
			gwHost,
			strings.Join(gwCmdOpts, " "),
		)

		cmdOpts = append(cmdOpts, "-o", proxyOpt)
	}

	cmdOpts = append(cmdOpts, a.ConnOpts.RawOpts...)

	return cmdOpts
}

func (a SSHArgs) gwOpts() (string, string, string) {
	if a.ConnOpts.GatewayDisable {
		return "", "", ""
	}

	// Take server provided gateway options
	username := a.Result.GatewayUsername
	host := a.Result.GatewayHost

	if len(a.ConnOpts.GatewayUsername) > 0 {
		username = a.ConnOpts.GatewayUsername
	}

	if len(a.ConnOpts.GatewayHost) > 0 {
		host = a.ConnOpts.GatewayHost
	}

	privKeyPath := a.ConnOpts.GatewayPrivateKeyPath

	return username, host, privKeyPath
}
