package ssh

import (
	"os/signal"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type Provider struct {
	streamingSSH ComboRunner
	resultsSSH   ComboRunner
	scp          ComboRunner
}

func NewProvider(cmdRunner boshsys.CmdRunner, fs boshsys.FileSystem, ui boshui.UI, logger boshlog.Logger) Provider {
	sshSessionFactory := func(connOpts ConnectionOpts, result boshdir.SSHResult) Session {
		return NewSessionImpl(connOpts, SessionImplOpts{ForceTTY: true}, result, fs)
	}

	streamingWriter := NewStreamingWriter(boshui.NewComboWriter(ui))

	streamingSSH := NewComboRunner(
		cmdRunner, sshSessionFactory, signal.Notify, streamingWriter, fs, ui, logger)

	resultsSSH := NewComboRunner(
		cmdRunner, sshSessionFactory, signal.Notify, NewResultsWriter(ui), fs, ui, logger)

	scpSessionFactory := func(connOpts ConnectionOpts, result boshdir.SSHResult) Session {
		return NewSessionImpl(connOpts, SessionImplOpts{}, result, fs)
	}

	scp := NewComboRunner(cmdRunner, scpSessionFactory, signal.Notify, streamingWriter, fs, ui, logger)

	return Provider{streamingSSH: streamingSSH, resultsSSH: resultsSSH, scp: scp}
}

func (p Provider) NewResultsSSHRunner(interactive bool) Runner {
	return NewNonInteractiveRunner(p.resultsSSH)
}

func (p Provider) NewSSHRunner(interactive bool) Runner {
	if interactive {
		return NewInteractiveRunner(p.streamingSSH)
	}
	return NewNonInteractiveRunner(p.streamingSSH)
}

func (p Provider) NewSCPRunner() SCPRunner { return NewSCPRunner(p.scp) }
