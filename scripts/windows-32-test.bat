git fetch
git checkout %GIT_COMMIT%
git submodule update --init

powershell -command set-executionpolicy remotesigned
powershell .\bin\replace-sha.ps1

SET GOPATH=c:\jenkins\workspace\go-cli-tests-windows32Bit
c:\Go\bin\go build -v -o cf-windows-386.exe main
c:\Go\bin\go test -i ./cf/...
c:\Go\bin\go test -v ./cf/...
