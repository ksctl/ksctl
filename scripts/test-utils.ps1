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
go test -fuzz=Fuzz -fuzztime 10s -v cloud_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v cni_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v name_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v storage_test.go utils.go
go test -fuzz=Fuzz -fuzztime 10s -v distro_test.go utils.go
go test . -v && Set-Location -
