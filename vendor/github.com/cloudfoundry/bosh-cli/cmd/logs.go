package cmd

import (
	"fmt"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
)

type LogsCmd struct {
	deployment      boshdir.Deployment
	downloader      Downloader
	uuidGen         boshuuid.Generator
	nonIntSSHRunner boshssh.Runner
}

func NewLogsCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	uuidGen boshuuid.Generator,
	nonIntSSHRunner boshssh.Runner,
) LogsCmd {
	return LogsCmd{
		deployment:      deployment,
		downloader:      downloader,
		uuidGen:         uuidGen,
		nonIntSSHRunner: nonIntSSHRunner,
	}
}

func (c LogsCmd) Run(opts LogsOpts) error {
	if opts.Follow || opts.Num > 0 {
		return c.tail(opts)
	}
	return c.fetch(opts)
}

func (c LogsCmd) tail(opts LogsOpts) error {
	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	result, err := c.deployment.SetUpSSH(opts.Args.Slug, sshOpts)
	if err != nil {
		return err
	}

	defer func() {
		_ = c.deployment.CleanUpSSH(opts.Args.Slug, sshOpts)
	}()

	err = c.nonIntSSHRunner.Run(connOpts, result, c.buildTailCmd(opts))
	if err != nil {
		return bosherr.WrapErrorf(err, "Running follow over non-interactive SSH")
	}

	return nil
}

func (c LogsCmd) buildTailCmd(opts LogsOpts) []string {
	cmd := []string{"sudo", "tail"}

	if opts.Follow {
		// -F for continuing to follow after renames
		cmd = append(cmd, "-F")
	}

	if opts.Num > 0 {
		cmd = append(cmd, "-n", strconv.Itoa(opts.Num))
	}

	if opts.Quiet {
		cmd = append(cmd, "-q")
	}

	var logsDir string

	if opts.Agent {
		logsDir = "/var/vcap/bosh/log"
	} else {
		logsDir = "/var/vcap/sys/log"
	}

	if len(opts.Jobs) > 0 {
		for _, job := range opts.Jobs {
			cmd = append(cmd, fmt.Sprintf("%s/%s/*.log", logsDir, job))
		}
	} else if len(opts.Filters) > 0 {
		for _, filter := range opts.Filters {
			cmd = append(cmd, fmt.Sprintf("%s/%s", logsDir, filter))
		}
	} else {
		// includes only directory and its subdirectories
		cmd = append(cmd, fmt.Sprintf("%s/{**/,}*.log", logsDir))
	}

	return cmd
}

func (c LogsCmd) fetch(opts LogsOpts) error {
	slug := opts.Args.Slug
	name := c.deployment.Name()

	if len(slug.Name()) > 0 {
		name += "." + slug.Name()
	}

	if len(slug.IndexOrID()) > 0 {
		name += "." + slug.IndexOrID()
	}

	result, err := c.deployment.FetchLogs(slug, opts.Filters, opts.Agent)
	if err != nil {
		return err
	}

	err = c.downloader.Download(
		result.BlobstoreID,
		result.SHA1,
		name,
		opts.Directory.Path,
	)
	if err != nil {
		return bosherr.WrapError(err, "Downloading logs")
	}

	return nil
}
