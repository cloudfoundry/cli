package options

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/options/multiopt"
	"github.com/pborman/getopt"
)

type TTYRequest int

const (
	REQUEST_TTY_AUTO TTYRequest = iota
	REQUEST_TTY_NO
	REQUEST_TTY_YES
	REQUEST_TTY_FORCE
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

	getoptSet                       *getopt.Set
	indexOption                     getopt.Option
	skipHostValidationOption        getopt.Option
	skipRemoteExecutionOption       getopt.Option
	disableTerminalAllocationOption getopt.Option
	forceTerminalAllocationOption   getopt.Option
	localForwardingOption           getopt.Option

	localForwardMultiVal *multiopt.MultiValue
}

var UsageError = errors.New("Invalid usage")

func NewSSHOptions() *SSHOptions {
	sshOptions := &SSHOptions{}

	opts := getopt.New()

	sshOptions.indexOption = opts.UintVarLong(
		&sshOptions.Index,
		"index",
		'i',
		"application instance index",
		"app-instance-index",
	)

	sshOptions.skipHostValidationOption = opts.BoolVarLong(
		&sshOptions.SkipHostValidation,
		"skip-host-validation",
		'k',
		"skip host key validation",
	).SetFlag()

	sshOptions.skipRemoteExecutionOption = opts.BoolVar(
		&sshOptions.SkipRemoteExecution,
		'N',
		"do not execute a remote command",
	).SetFlag()

	var force, disable bool
	sshOptions.forceTerminalAllocationOption = opts.BoolVar(&force, 't', "force pseudo-tty allocation").SetFlag()
	sshOptions.disableTerminalAllocationOption = opts.BoolVar(&disable, 'T', "disable pseudo-tty allocation").SetFlag()

	sshOptions.localForwardMultiVal = &multiopt.MultiValue{}
	sshOptions.localForwardingOption = opts.Var(
		sshOptions.localForwardMultiVal,
		'L',
		"local port forward specification",
		"[bind_address:]port:host:hostport",
	)

	sshOptions.getoptSet = opts

	return sshOptions
}

func (o *SSHOptions) Parse(args []string) error {
	opts := o.getoptSet
	err := opts.Getopt(args, nil)
	if err != nil {
		return err
	}

	if len(args) == 0 || args[0] != "ssh" {
		return UsageError
	}

	if opts.NArgs() == 0 {
		return UsageError
	}

	o.AppName = opts.Arg(0)

	if opts.NArgs() > 0 {
		err = opts.Getopt(opts.Args(), nil)
		if err != nil {
			return err
		}

		o.Command = opts.Args()
	}

	if o.localForwardingOption.Seen() {
		for _, arg := range o.localForwardMultiVal.Values() {
			forwardSpec, err := o.parseLocalForwardingSpec(arg)
			if err != nil {
				return err
			}
			o.ForwardSpecs = append(o.ForwardSpecs, *forwardSpec)
		}
	}

	if o.forceTerminalAllocationOption.Count() == 1 {
		o.TerminalRequest = REQUEST_TTY_YES
	} else if o.forceTerminalAllocationOption.Count() > 1 {
		o.TerminalRequest = REQUEST_TTY_FORCE
	}

	if o.disableTerminalAllocationOption.Count() != 0 {
		o.TerminalRequest = REQUEST_TTY_NO
	}

	return nil
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

func SSHUsage() string {
	b := &bytes.Buffer{}

	o := NewSSHOptions()
	o.getoptSet.SetProgram("ssh")
	o.getoptSet.SetParameters("app-name [command]")
	o.getoptSet.PrintUsage(b)

	return b.String()
}
