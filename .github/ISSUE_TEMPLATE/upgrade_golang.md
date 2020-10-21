## Phase 1: Confirm that upgrade works locally:

- [ ] Read the release notes: https://golang.org/doc/devel/release.html
- [ ] Update [`install-ubuntu.sh`'s GoLang version](https://github.com/cloudfoundry/cli-workstation/blob/master/install-ubuntu.sh#L4) and rerun on local workstation.
  - [ ] Update the GOROOT variable in [`0000-paths.bash`](https://github.com/cloudfoundry/cli-workstation/blob/master/dotfiles/bashit_custom_linux/0000-paths.bash#L9)
  - [ ] This may require rerunning the `GoUpdateBinaries` in `vim` on all the workstations after GoLang was updated.
- [ ] Run the unit and integration tests. Fix any tests that fail and commit changes.
- [ ] Run `go get -u golang.org/x/<pkg>` for `golang.org/x/` packages we directly depend on
  - [ ] Make sure to rerun the unit and integration tests.
- [ ] Update the version number in the [developer guide](https://github.com/cloudfoundry/cli/blob/master/.github/CONTRIBUTING.md#development-environment-setup).
- [ ] Update the [version in `.travis.yml`](https://github.com/cloudfoundry/cli/blob/master/.travis.yml#L3) Verify the change by looking at the travis output for a PR or a commit after the change to `.travis.yml` is pushed.
- [ ] Commit and Push the `install-ubuntu.sh` version update, and rerun it on all the workstations.
- [ ] Individually pause all the jobs in the `cli` group in Concourse.
- [ ] Push changes made to cf cli repo.

## Phase 2: Update Pipeline:
- [ ] Update and push the [`ci/cli-base/Dockerfile` image's `FROM golang:<major>.<minor>`](https://github.com/cloudfoundry/cli/blob/master/ci/cli-base/Dockerfile#L1) . Then run the `create-cli-base-image` Concourse Job.
- [ ] Upgrade GoLang on the OSX Worker: Run `brew update` and `brew upgrade golang`. Connection instructions [here](https://github.com/cloudfoundry/cli-private/blob/master/mac-worker-setup.md)
- [ ] Upgrade GoLang on the Windows worker(s) by updating [this pipeline](https://ci.cli.fun/teams/main/pipelines/concourse)
- [ ] Update TARGET_GO_VERSION in `cli-ci` with the new version #
- [ ] Refly the pipelines with `ci/bin/reconfigure-pipelines`, `ci/bin/reconfigure-v7-pipelines` and `ci/bin/reconfigure-v8-pipelines`
- [ ] Starting from the unit tests, unpause each job - waiting for the job to pass with the latest version of GoLang before unpausing the subsequent job.
