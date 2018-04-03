package releasedir

import (
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FSGitRepo struct {
	// Will use CmdRunner's WorkingDir when exec-ing to avoid --git-dir weirdness
	dirPath string
	runner  boshsys.CmdRunner
	fs      boshsys.FileSystem
}

func NewFSGitRepo(dirPath string, runner boshsys.CmdRunner, fs boshsys.FileSystem) FSGitRepo {
	return FSGitRepo{dirPath: dirPath, runner: runner, fs: fs}
}

func (r FSGitRepo) Init() error {
	_, _, _, err := r.runner.RunCommand("git", "init", r.dirPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Initing git repo")
	}

	ignoreTpl := `config/private.yml
blobs
dev_releases
releases/*.tgz
releases/**/*.tgz
.dev_builds
.final_builds/jobs/**/*.tgz
.final_builds/packages/**/*.tgz
.DS_Store
.idea
*.swp
*~
*#
#*
`

	err = r.fs.WriteFileString(filepath.Join(r.dirPath, ".gitignore"), ignoreTpl)
	if err != nil {
		return bosherr.WrapError(err, "Creating .gitignore file")
	}

	return nil
}

func (r FSGitRepo) LastCommitSHA() (string, error) {
	cmd := boshsys.Command{
		Name:       "git",
		Args:       []string{"rev-parse", "--short", "HEAD"},
		WorkingDir: r.dirPath,
	}
	stdout, stderr, _, err := r.runner.RunComplexCommand(cmd)
	if err != nil {
		if r.isNotGitRepo(stderr) {
			return "non-git", nil
		}

		if strings.Contains(stderr, "Needed a single revision") {
			return "empty", nil
		}

		return "", bosherr.WrapErrorf(err, "Checking last commit SHA")
	}

	return strings.TrimSpace(stdout), nil
}

func (r FSGitRepo) MustNotBeDirty(force bool) (bool, error) {
	cmd := boshsys.Command{
		Name:       "git",
		Args:       []string{"status", "--short"},
		WorkingDir: r.dirPath,
	}
	stdout, stderr, _, err := r.runner.RunComplexCommand(cmd)
	if err != nil {
		if r.isNotGitRepo(stderr) {
			return false, nil
		}

		return false, bosherr.WrapErrorf(err, "Checking status")
	}

	// Strip out newline which is added if there are any changes
	stdout = strings.TrimSpace(stdout)

	if len(stdout) > 0 {
		if force {
			return true, nil
		}
		return true, bosherr.Errorf("Git repository has local modifications:\n\n%s", stdout)
	}

	return false, nil
}

func (r FSGitRepo) isNotGitRepo(stderr string) bool {
	return strings.Contains(stderr, "Not a git repository")
}
