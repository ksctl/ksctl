#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
  Write-Output "PowerShell 5 or later is required to run Datree."
  Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
  break
}

Set-Location .\..\internal

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/cloudproviders/civo)"
Write-Output "--------------------------------------------"

Set-Location cloudproviders\civo
go test -tags testing_civo ./... -v && Set-Location -

