package application

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/requirements"
	sshCmd "github.com/cloudfoundry/cli/cf/ssh"
	"github.com/cloudfoundry/cli/cf/ssh/options"
	"github.com/cloudfoundry/cli/cf/ssh/terminal"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/simonleung8/flags"
	"github.com/simonleung8/flags/flag"
)

type SSH struct {
	ui            terminal.UI
	config        core_config.Reader
	gateway       net.Gateway
	appReq        requirements.ApplicationRequirement
	sshCodeGetter commands.SSHCodeGetter
	opts          *options.SSHOptions
	secureShell   sshCmd.SecureShell
}

type sshInfo struct {
	SSHEndpoint            string `json:"app_ssh_endpoint"`
	SSHEndpointFingerprint string `json:"app_ssh_host_key_fingerprint"`
}

func init() {
	command_registry.Register(&SSH{})
}

func (cmd *SSH) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["L"] = &cliFlags.StringSliceFlag{Name: "L", Usage: T("Local port forward specification. This flag can be defined more than once.")}
	fs["command"] = &cliFlags.StringSliceFlag{Name: "command", ShortName: "c", Usage: T("Command to run. This flag can be defined more than once.")}
	fs["app-instance-index"] = &cliFlags.IntFlag{Name: "app-instance-index", ShortName: "i", Usage: T("Application instance index")}
	fs["skip-host-validation"] = &cliFlags.BoolFlag{Name: "skip-host-validation", ShortName: "k", Usage: T("Skip host key validation")}
	fs["skip-remote-execution"] = &cliFlags.BoolFlag{Name: "skip-remote-execution", ShortName: "N", Usage: T("Do not execute a remote command")}
	fs["request-pseudo-tty"] = &cliFlags.BoolFlag{Name: "request-pseudo-tty", ShortName: "t", Usage: T("Request pseudo-tty allocation")}
	fs["force-pseudo-tty"] = &cliFlags.BoolFlag{Name: "force-pseudo-tty", ShortName: "tt", Usage: T("Force pseudo-tty allocation")}
	fs["disable-pseudo-tty"] = &cliFlags.BoolFlag{Name: "disable-pseudo-tty", ShortName: "T", Usage: T("Disable pseudo-tty allocation")}

	return command_registry.CommandMetadata{
		Name:        "ssh",
		Description: T("SSH to an application container instance"),
		Usage:       T("CF_NAME ssh APP_NAME [-i app-instance-index] [-c command] [-L [bind_address:]port:host:hostport] [--skip-host-validation] [--skip-remote-execution] [--request-pseudo-tty] [--force-pseudo-tty] [--disable-pseudo-tty]"),
		Flags:       fs,
	}
}

func (cmd *SSH) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) == 0 {
		cmd.ui.Failed(T("Incorrect Usage. Requires APP_NAME as argument") + "\n\n" + command_registry.Commands.CommandUsage("ssh"))
	}

	if fc.IsSet("i") && fc.Int("i") < 0 {
		cmd.ui.Failed(fmt.Sprintf(T("Incorrect Usage:")+" %s\n\n%s", T("Value for flag 'app-instance-index' cannot be negative"), command_registry.Commands.CommandUsage("ssh")))
	}

	cmd.opts, err = options.NewSSHOptions(fc)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Incorrect Usage:")+" %s\n\n%s", err.Error(), command_registry.Commands.CommandUsage("ssh")))
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(cmd.opts.AppName)
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return
}

func (cmd *SSH) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.gateway = deps.Gateways["cloud-controller"]

	if deps.WilecardDependency != nil {
		cmd.secureShell = deps.WilecardDependency.(sshCmd.SecureShell)
	}

	//get get-ssh-code for dependency
	sshCodeGetter := command_registry.Commands.FindCommand("get-ssh-code")
	sshCodeGetter = sshCodeGetter.SetDependency(deps, false)
	cmd.sshCodeGetter = sshCodeGetter.(commands.SSHCodeGetter)

	return cmd
}

func (cmd *SSH) Execute(fc flags.FlagContext) {
	app := cmd.appReq.GetApplication()
	info, err := cmd.getSSHEndpointInfo()
	if err != nil {
		cmd.ui.Failed(T("Error getting SSH info:") + err.Error())
	}

	sshAuthCode, err := cmd.sshCodeGetter.Get()
	if err != nil {
		cmd.ui.Failed(T("Error getting one time auth code: ") + err.Error())
	}

	//init secureShell if it is not already set by SetDependency() with fakes
	if cmd.secureShell == nil {
		cmd.secureShell = sshCmd.NewSecureShell(
			sshCmd.DefaultSecureDialer(),
			sshTerminal.DefaultHelper(),
			sshCmd.DefaultListenerFactory(),
			30*time.Second,
			app,
			info.SSHEndpointFingerprint,
			info.SSHEndpoint,
			sshAuthCode,
		)
	}

	err = cmd.secureShell.Connect(cmd.opts)
	if err != nil {
		cmd.ui.Failed(T("Error opening SSH connection: ") + err.Error())
	}
	defer cmd.secureShell.Close()

	err = cmd.secureShell.LocalPortForward()
	if err != nil {
		cmd.ui.Failed(T("Error forwarding port: ") + err.Error())
	}

	if cmd.opts.SkipRemoteExecution {
		err = cmd.secureShell.Wait()
	} else {
		err = cmd.secureShell.InteractiveSession()
	}

	if err == nil {
		return
	}

	if exitError, ok := err.(*ssh.ExitError); ok {
		exitStatus := exitError.ExitStatus()
		if sig := exitError.Signal(); sig != "" {
			cmd.ui.Say(fmt.Sprintf(T("Process terminated by signal: %s. Exited with")+" %d.\n", sig, exitStatus))
		}
		os.Exit(exitStatus)
	} else {
		cmd.ui.Failed(T("Error: ") + err.Error())
	}
}

func (cmd *SSH) getSSHEndpointInfo() (sshInfo, error) {
	info := sshInfo{}
	apiErr := cmd.gateway.GetResource(cmd.config.ApiEndpoint()+"/v2/info", &info)
	return info, apiErr
}
