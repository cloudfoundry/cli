!include "env_var_update.nsh"

Name "CloudFoundry CLI"
OutFile "cf_installer.exe"

InstallDir $PROGRAMFILES\CloudFoundry
InstallDirRegKey HKLM "Software\CloudFoundryCLI" "Install_Dir"

RequestExecutionLevel admin

Page directory
Page instfiles

; The stuff to install
Section "CloudFoundry CLI (required)"

  SectionIn RO
  
  ; Set output path to the installation directory.
  SetOutPath $INSTDIR
  
  ; Put file there
  File "cf.exe"
  
  ; Write the installation path into the registry
  WriteRegStr HKLM Software\CloudFoundryCLI "Install_Dir" "$INSTDIR"
  
  ; Add output directory to system path
  ${EnvVarUpdate} $0 "PATH" "A" "HKLM" "$INSTDIR"

SectionEnd
