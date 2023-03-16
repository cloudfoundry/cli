# This image is not being used by the github actions workflow
# because gh-actions doesn't support windows based images
# Keeping this file as we expect to use it in the future
FROM mcr.microsoft.com/windows/servercore:ltsc2019
SHELL ["powershell.exe"]

ARG CHOCO_THUMBPRINT=83AC7D88C66CB8680BCE802E0F0F5C179722764B
RUN mkdir \setup

RUN (New-Object System.Net.WebClient).DownloadFile('https://chocolatey.org/install.ps1', '\setup\installChocolatey.ps1')
RUN (Get-AuthenticodeSignature \setup\installChocolatey.ps1).SignerCertificate.Thumbprint > \setup\thumbprint
RUN if ((type \setup\thumbprint) -ne $env:CHOCO_THUMBPRINT) { \
      throw 'chocolatey installer thumbprint does not match expected. see https://docs.chocolatey.org/en-us/information/security' \
    }
RUN \setup\installChocolatey.ps1
RUN Remove-Item -Recurse \setup


RUN choco install --no-progress -r -y innosetup
