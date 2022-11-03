#Requires -Version 5

$old_erroractionpreference = $erroractionpreference
$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run Datree."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

Write-Host "Welcome to Installation" -ForegroundColor DarkGreen

mkdir -p $env:HOME\.ksctl\cred
mkdir -p $env:HOME\.ksctl\config\civo
mkdir -p $env:HOME\.ksctl\config\local

New-Item $env:HOME\.ksctl\cred\civo
New-Item $env:HOME\.ksctl\cred\aws
New-Item $env:HOME\.ksctl\cred\azure

Set-Variable $env:GOOS = 'windows'
Set-Variable $env:GOARCH = 'amd64'

Set-Location .\src\cli\
go build -v -o ksctl.exe .

#Move-Item ksctl.exe $env:HOME\.ksctl\

$localAppDataPath = $env:LOCALAPPDATA
$ksctl = Join-Path "$localAppDataPath" 'ksctl'

Write-Information "Path of AppDataPath $ksctl"

New-Item -ItemType Directory -Force -Path $ksctl | Out-Null

Copy-Item ksctl.exe -Destination "$ksctl/" -Force | Out-Null


Write-Host "[V] Finished Installation" -ForegroundColor DarkGreen
Write-Host ""
Write-Host "To run datree globally, please follow these steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "    1. Run the following command as administrator: ``setx PATH `"`$env:path;$ksctl`" -m``"
