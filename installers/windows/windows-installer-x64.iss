[Setup]
ChangesEnvironment=yes
AppName=Cloud Foundry CLI
AppVersion=VERSION
AppVerName=Cloud Foundry CLI version VERSION
DefaultDirName={pf}\CloudFoundry
AppPublisher=Cloud Foundry Foundation
SignTool=signtool sign /f $qSIGNTOOL_CERT_PATH$q /p $qSIGNTOOL_CERT_PASSWORD$q /t http://timestamp.comodoca.com/authenticode $f
ArchitecturesInstallIn64BitMode=x64 ia64
ArchitecturesAllowed=x64 ia64

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: Uninstall32Bit() and NeedsAddPath(ExpandConstant('{app}'))

[Files]
Source: CF_SOURCE; DestDir: "{app}"

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

function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE,
    'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
    'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  // look for the path with leading and trailing semicolon
  // Pos() returns 0 if not found
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
