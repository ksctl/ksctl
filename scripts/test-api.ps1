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

Set-Location helpers
go test -fuzz=Fuzz -fuzztime 10s -v cloud_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v cni_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v name_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v storage_test.go fields.go
go test -fuzz=Fuzz -fuzztime 10s -v distro_test.go fields.go
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


Write-Output "--------------------------------------------"
Write-Output "|   Testing (internal/storage/local)"
Write-Output "--------------------------------------------"

Set-Location storage\local
go test . -v && Set-Location -


Write-Output "-------------------------------------------------"
Write-Output "|   Testing (ksctl-components\agent)"
Write-Output "-------------------------------------------------"

Set-Location ksctl-components\agent
go test . -v && Set-Location -

# Write-Output "-------------------------------------------------"
# Write-Output "|   Testing (internal/storage/external/mongodb)"
# Write-Output "-------------------------------------------------"
#
# Set-Location storage\external\mongodb
# go test . -v && Set-Location -
