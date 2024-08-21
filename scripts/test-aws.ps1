#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
  Write-Output "PowerShell 5 or later is required to run Datree."
  Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
  break
}

Set-Location .\..\internal\

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/cloudproviders/aws)"
Write-Output "--------------------------------------------"

Set-Location cloudproviders\aws
go test -tags testing_aws ./... -v && Set-Location -
