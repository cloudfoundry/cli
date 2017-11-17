package dialcontext

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/anynines/go-proxy-setup-ntlm/proxysetup/ntlm"
)

// Go does not support NTLM proxy authentication by default.
//
// An attempt https://github.com/golang/go/issues/22288 to add NTLM proxy
// authentication to Go's code base has not been accepted.
//
// But there is a workaround/hack overwriting http.Transport.DialContext to do
// NTLM proxy authentication.

// Returns the passed DialContext by default.
//
// Experimental:
// Returns NTLM proxy authentication handler if NTLM_PROXY is set.
// The environment variable NTLM_PROXY contains the proxy to be used. Works on
// Windows only.
func FromEnvironment(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) func(ctx context.Context, network, addr string) (net.Conn, error) {
	if len(ntlmProxyAddr.Get()) > 0 {
		return func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialContext(ctx, network, ntlmProxyAddr.Get())
			if err != nil {
				return conn, err
			}
			err = ntlm.ProxySetup(conn, address)
			if err != nil {
				return conn, err
			}
			return conn, err
		}
	}

	return dialContext
}

//
// Private implementation past this point.
//

var (
	ntlmProxyAddr = &envOnce{
		names: []string{"NTLM_PROXY", "ntlm_proxy"},
	}
)

// envOnce looks up an environment variable (optionally by multiple
// names) once. It mitigates expensive lookups on some platforms
// (e.g. Windows).
type envOnce struct {
	names []string
	once  sync.Once
	val   string
}

func (e *envOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envOnce) init() {
	for _, n := range e.names {
		e.val = os.Getenv(n)
		if e.val != "" {
			return
		}
	}
}

// reset is used by tests
func (e *envOnce) reset() {
	e.once = sync.Once{}
	e.val = ""
}
