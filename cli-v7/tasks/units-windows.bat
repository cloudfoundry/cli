SET TARGET_V7=true
SET GOPATH=%CD%\gopath
SET PATH=C:\Go\bin;C:\Program Files\Git\cmd\;%GOPATH%\bin;%PATH%

md %GOPATH%\src\code.cloudfoundry.org\
mklink /D %CD%\cli %GOPATH%\src\code.cloudfoundry.org\cli

cd %GOPATH%\src\code.cloudfoundry.org\cli

powershell -command set-executionpolicy remotesigned

go version

go get -u github.com/onsi/ginkgo/ginkgo

ginkgo version

ginkgo -r -randomizeAllSpecs -randomizeSuites -skipPackage integration -flakeAttempts=2 -tags="V7" .
