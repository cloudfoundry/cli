package ssh

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type SessionImpl struct {
	connOpts ConnectionOpts
	sessOpts SessionImplOpts
	result   boshdir.SSHResult

	privKeyFile    boshsys.File
	knownHostsFile boshsys.File

	fs boshsys.FileSystem
}

type SessionImplOpts struct {
	ForceTTY bool
}

func NewSessionImpl(
	connOpts ConnectionOpts,
	sessOpts SessionImplOpts,
	result boshdir.SSHResult,
	fs boshsys.FileSystem,
) *SessionImpl {
	return &SessionImpl{connOpts: connOpts, sessOpts: sessOpts, result: result, fs: fs}
}

func (r *SessionImpl) Start() (SSHArgs, error) {
	var err error

	r.privKeyFile, err = r.makePrivKeyFile()
	if err != nil {
		return SSHArgs{}, err
	}

	r.knownHostsFile, err = r.makeKnownHostsFile()
	if err != nil {
		_ = r.fs.RemoveAll(r.privKeyFile.Name())
		return SSHArgs{}, err
	}

	args := NewSSHArgs(
		r.connOpts,
		r.result,
		r.sessOpts.ForceTTY,
		r.privKeyFile,
		r.knownHostsFile,
	)

	return args, nil
}

func (r *SessionImpl) Finish() error {
	// Make sure to try to delete all files regardless of errors
	privKeyErr := r.fs.RemoveAll(r.privKeyFile.Name())
	knownHostsErr := r.fs.RemoveAll(r.knownHostsFile.Name())

	if privKeyErr != nil {
		return privKeyErr
	}

	if knownHostsErr != nil {
		return knownHostsErr
	}

	return nil
}

func (r SessionImpl) makePrivKeyFile() (boshsys.File, error) {
	file, err := r.fs.TempFile("ssh-priv-key")
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating temp file for SSH private key")
	}

	_, err = file.Write([]byte(r.connOpts.PrivateKey))
	if err != nil {
		_ = r.fs.RemoveAll(file.Name())
		return nil, bosherr.WrapErrorf(err, "Writing SSH private key")
	}

	return file, nil
}

func (r SessionImpl) makeKnownHostsFile() (boshsys.File, error) {
	file, err := r.fs.TempFile("ssh-known-hosts")
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating temp file for SSH known hosts")
	}

	var content string

	for _, host := range r.result.Hosts {
		if len(host.HostPublicKey) > 0 {
			content += fmt.Sprintf("%s %s\n", host.Host, host.HostPublicKey)
		}
	}

	if len(content) > 0 {
		_, err := file.Write([]byte(content))
		if err != nil {
			_ = r.fs.RemoveAll(file.Name())
			return nil, bosherr.WrapErrorf(err, "Writing SSH known hosts")
		}
	}

	return file, nil
}
