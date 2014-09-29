git fetch
git checkout %GIT_COMMIT%

go get github.com/jteeuwen/go-bindata/...
go-bindata -pkg resources -ignore ".go" -o cf/resources/i18n_resources.go cf/i18n/resources/... cf/i18n/test_fixtures/...

powershell -command set-executionpolicy remotesigned
powershell .\bin\replace-sha.ps1

$env:GOPATH = "%CD%\Godeps\_workspace;c:\Users\Administrator\go"
Get-ChildItem Env:GOPATH

c:\Go\bin\go build -v -o cf-windows-386.exe ./main
c:\Go\bin\go test -i ./cf/... ./generic/... ./testhelpers/... ./main/...
c:\Go\bin\go test -cover -v ./cf/... ./generic/... ./testhelpers/... ./main/...
