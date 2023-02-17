#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run Datree."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

Write-Host "Welcome to Installation" -ForegroundColor DarkGreen

mkdir -Force $env:USERPROFILE\.ksctl\cred
mkdir -Force $env:USERPROFILE\.ksctl\config\civo
mkdir -Force $env:USERPROFILE\.ksctl\config\civo\ha
mkdir -Force $env:USERPROFILE\.ksctl\config\civo\managed
mkdir -Force $env:USERPROFILE\.ksctl\config\local
mkdir -Force $env:USERPROFILE\.ksctl\config\azure
mkdir -Force $env:USERPROFILE\.ksctl\config\azure\ha
mkdir -Force $env:USERPROFILE\.ksctl\config\azure\managed

$env:GOOS = 'windows'
$env:GOARCH = 'amd64'

Set-Location .\cli\
go build -v -o ksctl.exe .

#Move-Item ksctl.exe $env:USERPROFILE\.ksctl\

$localAppDataPath = $env:LOCALAPPDATA
$ksctl = Join-Path "$localAppDataPath" 'ksctl'

Write-Information "Path of AppDataPath $ksctl"

New-Item -ItemType Directory -Force -Path $ksctl | Out-Null

Copy-Item ksctl.exe -Destination "$ksctl/" -Force | Out-Null

Remove-Item ksctl.exe

Set-Location ..\.. | Out-Null

Write-Host "[V] Finished Installation" -ForegroundColor DarkGreen
Write-Host ""
Write-Host "To run ksctl globally, please follow these steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "    1. Run the following command as administrator: ``setx PATH `"`$env:path;$ksctl`" -m``"
