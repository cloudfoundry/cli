[Setup]
ChangesEnvironment=yes
AlwaysShowDirOnReadyPage=yes
AppName=Cloud Foundry CLI
AppVersion=VERSION
AppVerName=Cloud Foundry CLI version VERSION
AppPublisher=Cloud Foundry Foundation
ArchitecturesInstallIn64BitMode=x64 ia64
ArchitecturesAllowed=x64 ia64
PrivilegesRequired=none
DefaultDirName={pf}\Cloud Foundry
SetupIconFile=cf.ico
UninstallDisplayIcon={app}\cf.ico

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: IsAdminLoggedOn and Uninstall32Bit() and NeedsAddPath(ExpandConstant('{app}'))
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: not IsAdminLoggedOn and Uninstall32Bit() and NeedsAddPath(ExpandConstant('{app}'))

[Files]
Source: LICENSE; DestDir: "{app}"
Source: NOTICE; DestDir: "{app}"
Source: cf8.exe; DestDir: "{app}"
Source: cf.ico; DestDir: "{app}"

[Run]
Filename: "{cmd}"; Parameters: "/C mklink ""{app}\cf.exe"" ""{app}\cf8.exe"""

[UninstallDelete]
Type: files; Name: "{app}\cf.exe"
Type: dirifempty; Name: "{app}"

[Code]
function Uninstall32Bit(): Boolean;
var
  resultCode: Integer;
  uninstallString: String;
  uninstallStringPath: String;
begin
  uninstallString := '';
  uninstallStringPath := 'SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Cloud Foundry CLI_is1';
  RegQueryStringValue(HKLM, uninstallStringPath, 'UninstallString', uninstallString);

  if uninstallString <> '' then
  begin
    uninstallString := RemoveQuotes(uninstallString);
    Exec(uninstallString, '/VERYSILENT /SUPPRESSMSGBOXES /NORESTART','', SW_HIDE, ewWaitUntilTerminated, resultCode)
  end;
  Result := true;
end;

#include "common.iss"
