$ErrorActionPreference = "Stop"
trap { $host.SetShouldExit(1) }

. "$PSScriptRoot\windows-setup.ps1"

pushd $Env:ROOT
  Set-ExecutionPolicy RemoteSigned -Scope Process

  go version
  ginkgo version

  ginkgo -r `
  -randomizeAllSpecs `
  -randomizeSuites `
  -skipPackage integration `
  -flakeAttempts=2 `
  -tags="V7"
popd
