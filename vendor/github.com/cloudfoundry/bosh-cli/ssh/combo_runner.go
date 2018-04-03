package ssh

import (
	"os"
	"syscall"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/hashicorp/go-multierror"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type ComboRunner struct {
	cmdRunner        boshsys.CmdRunner
	sessionFactory   func(ConnectionOpts, boshdir.SSHResult) Session
	signalNotifyFunc func(chan<- os.Signal, ...os.Signal)

	writer Writer
	fs     boshsys.FileSystem
	ui     boshui.UI

	logTag string
	logger boshlog.Logger
}

func NewComboRunner(
	cmdRunner boshsys.CmdRunner,
	sessionFactory func(ConnectionOpts, boshdir.SSHResult) Session,
	signalNotifyFunc func(chan<- os.Signal, ...os.Signal),
	writer Writer,
	fs boshsys.FileSystem,
	ui boshui.UI,
	logger boshlog.Logger,
) ComboRunner {
	return ComboRunner{
		cmdRunner:        cmdRunner,
		sessionFactory:   sessionFactory,
		signalNotifyFunc: signalNotifyFunc,

		writer: writer,
		fs:     fs,
		ui:     ui,

		logTag: "ComboRunner",
		logger: logger,
	}
}

func (r ComboRunner) Run(connOpts ConnectionOpts, result boshdir.SSHResult, cmdFactory func(boshdir.Host, SSHArgs) boshsys.Command) error {
	sess := r.sessionFactory(connOpts, result)

	sshArgs, err := sess.Start()
	if err != nil {
		return bosherr.WrapErrorf(err, "Setting up SSH session")
	}

	defer func() {
		_ = sess.Finish()
	}()

	cancelCh := make(chan struct{}, 1)

	go r.setUpInterrupt(cancelCh, sess)

	cmds := r.makeCmds(result.Hosts, sshArgs, cmdFactory)

	ps, doneCh := r.runCmds(cmds)

	return r.waitProcs(ps, doneCh, cancelCh)
}

type comboRunnerCmd struct {
	boshsys.Command
	InstanceWriter
}

func (r ComboRunner) makeCmds(hosts []boshdir.Host, sshArgs SSHArgs, cmdFactory func(boshdir.Host, SSHArgs) boshsys.Command) []comboRunnerCmd {
	var cmds []comboRunnerCmd

	for _, host := range hosts {
		cmd := cmdFactory(host, sshArgs)

		jobName := "?"
		if len(host.Job) > 0 {
			jobName = host.Job
		}

		instWriter := r.writer.ForInstance(jobName, host.IndexOrID)

		if cmd.Stdout == nil && cmd.Stderr == nil {
			cmd.Stdout = instWriter.Stdout()
			cmd.Stderr = instWriter.Stderr()
		}

		cmds = append(cmds, comboRunnerCmd{cmd, instWriter})
	}

	return cmds
}

func (r ComboRunner) runCmds(cmds []comboRunnerCmd) ([]boshsys.Process, chan []boshsys.Result) {
	var processes []boshsys.Process

	allResultsCh := make(chan boshsys.Result, len(cmds))

	for _, cmd := range cmds {
		process, err := r.cmdRunner.RunComplexCommandAsync(cmd.Command)
		if err != nil {
			r.logger.Error(r.logTag, "Process immediately failed")
			cmd.InstanceWriter.End(0, err)
			allResultsCh <- boshsys.Result{Error: err}
			continue
		}

		processes = append(processes, process)

		// Call Wait outside of goroutine
		// to make sure TerminateNicely is not called before
		resultCh := process.Wait()

		// local variable to keep it in scope
		instWriter := cmd.InstanceWriter

		go func() {
			result := <-resultCh
			instWriter.End(result.ExitStatus, result.Error)
			allResultsCh <- result
		}()
	}

	r.logger.Debug(r.logTag, "Started all processes")

	doneCh := make(chan []boshsys.Result)

	go func() {
		var rs []boshsys.Result

		for i := 0; i < len(cmds); i++ {
			rs = append(rs, <-allResultsCh)
		}

		doneCh <- rs
	}()

	return processes, doneCh
}

func (r ComboRunner) waitProcs(ps []boshsys.Process, doneCh chan []boshsys.Result, cancelCh chan struct{}) error {
	r.logger.Debug(r.logTag, "Waiting for all processes or cancel signal")

	for {
		select {
		case results := <-doneCh:
			var errs error

			for _, r := range results {
				if r.Error != nil {
					errs = multierror.Append(errs, r.Error)
				}
			}

			r.logger.Debug(r.logTag, "All processes finished '%#v' with errors '%s'", results, errs)

			r.writer.Flush()

			return errs

		case <-cancelCh:
			r.logger.Debug(r.logTag, "Received cancel signal")

			for _, p := range ps {
				err := p.TerminateNicely(10 * time.Second)
				if err != nil {
					r.logger.Error(r.logTag, "Failed to terminate with error '%s'", err.Error())
				}
			}

			// Expecting that after terminating all processes
			// doneCh will be signaled at some point.
		}
	}
}

func (r ComboRunner) setUpInterrupt(cancelCh chan<- struct{}, sess Session) {
	signalCh := make(chan os.Signal, 1)

	r.signalNotifyFunc(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	for sig := range signalCh {
		r.logger.Debug(r.logTag, "Received a signal: %v", sig)

		r.ui.PrintLinef("\nReceived a signal, exiting...\n")

		// Aggressively clear session, even though it may be cleared later
		_ = sess.Finish()

		cancelCh <- struct{}{}
	}
}
