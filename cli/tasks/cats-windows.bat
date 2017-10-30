SET GOPATH=%CD%\gopath
SET CATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests
SET CF_DIAL_TIMEOUT=15

SET PATH=C:\Go\bin;%PATH%
SET PATH=C:\Program Files\Git\cmd\;%PATH%
SET PATH=%GOPATH%\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files\cURL\bin;%PATH%
SET PATH=C:\Program Files\CloudFoundry;%PATH%
SET PATH=%CD%;%PATH%

SET CONFIG=%CD%\cats-config\integration_config.json

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
popd

SET CF_PLUGIN_HOME=%CD%
.\cf add-plugin-repo CATS-Test-CF-Community https://plugins.cloudfoundry.org
.\cf install-plugin -f -r CATS-Test-CF-Community "network-policy"
.\cf remove-plugin-repo CATS-Test-CF-Community

mkdir %CATSPATH%
xcopy /q /e /s cf-acceptance-tests %CATSPATH%

cd %CATSPATH%

ginkgo.exe -flakeAttempts=2 -slowSpecThreshold=180 -skip="go makes the app reachable via its bound route|SSO|takes effect after a restart, not requiring a push|doesn't die when printing 32MB|exercises basic loggregator|firehose data|dotnet-core|transparently proxies both reserved|manage droplet bits for an app" -nodes=%NODES%
