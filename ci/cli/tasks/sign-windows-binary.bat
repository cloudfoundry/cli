SET PATH=C:\Program Files (x86)\Windows Kits\10\bin\x64;%PATH%

signtool sign /f %SIGNTOOL_CERT_PATH% /p %SIGNTOOL_CERT_PASSWORD% /t http://timestamp.comodoca.com/authenticode %CD%\windows-signing\cf-cli_win32.exe
signtool sign /f %SIGNTOOL_CERT_PATH% /p %SIGNTOOL_CERT_PASSWORD% /t http://timestamp.comodoca.com/authenticode %CD%\windows-signing\cf-cli_winx64.exe

MOVE %CD%\windows-signing\*  %CD%\signed
