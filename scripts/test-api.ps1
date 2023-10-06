#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
  Write-Output "PowerShell 5 or later is required to run Datree."
  Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
  break
}

Set-Location .\..\pkg\

Write-Output "-----------------------------------"
Write-Output "|   Testing (pkg/utils)"
Write-Output "-----------------------------------"

Set-Location utils
go test . -v && Set-Location -

Write-Output "-----------------------------------"
Write-Output "|   Testing (pkg/logger)"
Write-Output "-----------------------------------"

Set-Location logger
go test . -v && Set-Location -

Set-Location .\..\internal

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/k8sdistros/k3s)"
Write-Output "--------------------------------------------"

Set-Location k8sdistros\k3s
go test . -v && Set-Location -

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/cloudproviders/local)"
Write-Output "--------------------------------------------"

Set-Location cloudproviders\local
go test . -v && Set-Location -

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/cloudproviders/civo)"
Write-Output "--------------------------------------------"

Set-Location cloudproviders\civo
go test . -v && Set-Location -

Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/cloudproviders/azure)"
Write-Output "--------------------------------------------"

Set-Location cloudproviders\azure
go test . -v && Set-Location -
