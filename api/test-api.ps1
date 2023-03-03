#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
  Write-Output "PowerShell 5 or later is required to run Datree."
  Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
  break
}

Write-Output "+-------------------------+"
Write-Output "|   Testing (api/util)    |"
Write-Output "+-------------------------+"

Set-Location utils
go test . -v && Set-Location -



Write-Output "+-------------------------+"
Write-Output "|   Testing (api/local)   |"
Write-Output "+-------------------------+"

Set-Location local
go test . -v && Set-Location -



Write-Output "+-------------------------+"
Write-Output "|   Testing (api/civo)    |"
Write-Output "+-------------------------+"

Set-Location civo
go test . -v && Set-Location -

Write-Output "+-------------------------+"
Write-Output "|   Testing (api/azure)    |"
Write-Output "+-------------------------+"

Set-Location azure
go test . -v && Set-Location -
