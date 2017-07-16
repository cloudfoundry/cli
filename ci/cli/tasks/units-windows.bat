SET GOPATH=%CD%\gopath
SET PATH=C:\Go\bin;C:\Program Files\Git\cmd\;%GOPATH%\bin;%PATH%

cd %GOPATH%\src\code.cloudfoundry.org\cli

powershell -command set-executionpolicy remotesigned

go version

go get github.com/onsi/ginkgo/ginkgo

ginkgo version

ginkgo -r -race -randomizeAllSpecs -randomizeSuites -skipPackage integration .
