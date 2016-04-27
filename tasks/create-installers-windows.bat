SET ROOT_DIR=%CD%
SET ESCAPED_ROOT_DIR=%ROOT_DIR:\=\\%
SET /p VERSION=<%ROOT_DIR%\cli\VERSION

SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files (x86)\Inno Setup 5;%PATH%

sed -i -e "s/VERSION/%VERSION%/" %ROOT_DIR%\cli\installers\windows\windows-installer.iss
sed -i -e "s/CF_SOURCE/%ESCAPED_ROOT_DIR%\\cf.exe/" %ROOT_DIR%\cli\installers\windows\windows-installer.iss

pushd %ROOT_DIR%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE cf-cli_winx64.exe ..\cf.exe
popd

ISCC %ROOT_DIR%\cli\installers\windows\windows-installer.iss

MOVE %ROOT_DIR%\cli\installers\windows\Output\setup.exe cf_installer.exe

zip %ROOT_DIR%\winstallers\cf-cli-installer_winx64.zip cf_installer.exe

pushd %ROOT_DIR%\cf-cli-binaries
	MOVE cf-cli_win32.exe ..\cf.exe
popd

ISCC %ROOT_DIR%\cli\installers\windows\windows-installer.iss

MOVE %ROOT_DIR%\cli\installers\windows\Output\setup.exe cf_installer.exe

zip %ROOT_DIR%\winstallers\cf-cli-installer_win32.zip cf_installer.exe
