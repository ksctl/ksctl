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
go test -fuzz=Fuzz -fuzztime 10s -v role_test.go fields.go
go test . -v && Set-Location -
