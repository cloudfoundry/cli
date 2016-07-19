package options

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/cf/flags"
)

type TTYRequest int

const (
	RequestTTYAuto TTYRequest = iota
	RequestTTYNo
	RequestTTYYes
	RequestTTYForce
)

type ForwardSpec struct {
	ListenAddress  string
	ConnectAddress string
}

type SSHOptions struct {
	AppName             string
	Command             []string
	Index               uint
	SkipHostValidation  bool
	SkipRemoteExecution bool
	TerminalRequest     TTYRequest
	ForwardSpecs        []ForwardSpec
}

func NewSSHOptions(fc flags.FlagContext) (*SSHOptions, error) {
	sshOptions := &SSHOptions{}

	sshOptions.AppName = fc.Args()[0]
	sshOptions.Index = uint(fc.Int("i"))
	sshOptions.SkipHostValidation = fc.Bool("k")
	sshOptions.SkipRemoteExecution = fc.Bool("N")
	sshOptions.Command = fc.StringSlice("c")

	if fc.IsSet("L") {
		for _, arg := range fc.StringSlice("L") {
			forwardSpec, err := sshOptions.parseLocalForwardingSpec(arg)
			if err != nil {
				return sshOptions, err
			}
			sshOptions.ForwardSpecs = append(sshOptions.ForwardSpecs, *forwardSpec)
		}
	}

	if fc.IsSet("t") && fc.Bool("t") {
		sshOptions.TerminalRequest = RequestTTYYes
	}

	if fc.IsSet("tt") && fc.Bool("tt") {
		sshOptions.TerminalRequest = RequestTTYForce
	}

	if fc.Bool("T") {
		sshOptions.TerminalRequest = RequestTTYNo
	}

	return sshOptions, nil
}

func (o *SSHOptions) parseLocalForwardingSpec(arg string) (*ForwardSpec, error) {
	arg = strings.TrimSpace(arg)

	parts := []string{}
	for remainder := arg; remainder != ""; {
		part, r, err := tokenizeForward(remainder)
		if err != nil {
			return nil, err
		}

		parts = append(parts, part)
		remainder = r
	}

	forwardSpec := &ForwardSpec{}
	switch len(parts) {
	case 4:
		if parts[0] == "*" {
			parts[0] = ""
		}
		forwardSpec.ListenAddress = fmt.Sprintf("%s:%s", parts[0], parts[1])
		forwardSpec.ConnectAddress = fmt.Sprintf("%s:%s", parts[2], parts[3])
	case 3:
		forwardSpec.ListenAddress = fmt.Sprintf("localhost:%s", parts[0])
		forwardSpec.ConnectAddress = fmt.Sprintf("%s:%s", parts[1], parts[2])
	default:
		return nil, fmt.Errorf("Unable to parse local forwarding argument: %q", arg)
	}

	return forwardSpec, nil
}

func tokenizeForward(arg string) (string, string, error) {
	switch arg[0] {
	case ':':
		return "", arg[1:], nil

	case '[':
		parts := strings.SplitAfterN(arg, "]", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("Argument missing closing bracket: %q", arg)
		}

		if parts[1][0] == ':' {
			return parts[0], parts[1][1:], nil
		}

		return "", "", fmt.Errorf("Unexpected token: %q", parts[1])

	default:
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) < 2 {
			return parts[0], "", nil
		}
		return parts[0], parts[1], nil
	}
}
