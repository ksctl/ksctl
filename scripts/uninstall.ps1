#Requires -Version 5

$old_erroractionpreference = $erroractionpreference
$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run Datree."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

Write-Host "Sorry to see you go" -ForegroundColor DarkGreen

$localAppDataPath = $env:LOCALAPPDATA
$ksctl = Join-Path "$localAppDataPath" 'ksctl'

Remove-Item -Force $ksctl | Out-Null
Remove-Item -Force $env:USERPROFILE\.ksctl | Out-Null

Write-Host "[V] Finished Uninstallation" -ForegroundColor DarkGreen
Write-Host ""
