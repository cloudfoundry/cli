git fetch
git checkout %GIT_COMMIT%
git submodule update --init

powershell -command set-executionpolicy remotesigned
powershell .\bin\replace-sha.ps1

SET GOPATH=c:\jenkins\workspace\go-cli-tests-windows64Bit
c:\Go\bin\go build -v -o cf-windows-amd64.exe main
c:\Go\bin\go test -i ./cf/...
c:\Go\bin\go test -v ./cf/...
