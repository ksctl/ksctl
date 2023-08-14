#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
  Write-Output "PowerShell 5 or later is required to run Datree."
  Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
  break
}

Write-Output "-----------------------------------"
Write-Output "|   Testing (api/utils)"
Write-Output "-----------------------------------"

Set-Location provider\utils
go test . -v && Set-Location -

Write-Output "-----------------------------------"
Write-Output "|   Testing (api/k8s_distro/k3s)"
Write-Output "-----------------------------------"

Set-Location k8s_distro\k3s
go test . -v && Set-Location -

Write-Output "-----------------------------------"
Write-Output "|   Testing (api/provider/local)"
Write-Output "-----------------------------------"

Set-Location provider\local
go test . -v && Set-Location -

Write-Output "-----------------------------------"
Write-Output "|   Testing (api/provider/civo)"
Write-Output "-----------------------------------"

Set-Location provider\civo
go test . -v && Set-Location -

Write-Output "-----------------------------------"
Write-Output "|   Testing (api/provider/azure)"
Write-Output "-----------------------------------"

Set-Location provider\azure
go test . -v && Set-Location -
