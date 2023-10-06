#Requires -Version 5

$erroractionpreference = 'stop' # quit if anything goes wrong

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Output "PowerShell 5 or later is required to run Datree."
    Write-Output "Upgrade PowerShell: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

Write-Host "Welcome to Installation" -ForegroundColor DarkGreen

mkdir -Force $env:USERPROFILE\.ksctl\cred
mkdir -Force $env:USERPROFILE\.ksctl\config\civo
mkdir -Force $env:USERPROFILE\.ksctl\config\civo\ha
mkdir -Force $env:USERPROFILE\.ksctl\config\civo\managed
mkdir -Force $env:USERPROFILE\.ksctl\config\azure
mkdir -Force $env:USERPROFILE\.ksctl\config\azure\ha
mkdir -Force $env:USERPROFILE\.ksctl\config\azure\managed
mkdir -Force $env:USERPROFILE\.ksctl\config\aws\ha
mkdir -Force $env:USERPROFILE\.ksctl\config\aws\managed
mkdir -Force $env:USERPROFILE\.ksctl\config\local
mkdir -Force $env:USERPROFILE\.ksctl\config\local\managed

Write-Host "Available Releases"  -ForegroundColor Cyan

$response = Invoke-RestMethod "https://api.github.com/repos/kubesimplify/ksctl/releases"
# get the release version

foreach ($release in $response.tag_name) {
    Write-Host "${release}" -ForegroundColor Cyan
}
$ReleaseVersion= Read-Host -Prompt "Enter the ksctl version to install"
$Arch= Read-Host -Prompt "Enter [1] for amd64 or x86_64 and [0] for arm64"

Set-Location $env:USERPROFILE

if ($Arch -eq 1) {
  $Arch="amd64"
} elseif ($Arch -eq 0) {
  $Arch="arm64"
} else {
  Write-Host "Invalid Arch" -ForegroundColor Red
  Exit 1
}

Invoke-WebRequest -Uri https://github.com/kubesimplify/ksctl/releases/download/v${ReleaseVersion}/ksctl_${ReleaseVersion}_checksums.txt -OutFile ksctl_${ReleaseVersion}_checksums.txt
Invoke-WebRequest -Uri https://github.com/kubesimplify/ksctl/releases/download/v${ReleaseVersion}/ksctl_${ReleaseVersion}_windows_${Arch}.tar.gz -OutFile ksctl_${ReleaseVersion}_windows_${Arch}.tar.gz
Invoke-WebRequest -Uri https://github.com/kubesimplify/ksctl/releases/download/v${ReleaseVersion}/ksctl_${ReleaseVersion}_windows_${Arch}.tar.gz.cert -OutFile ksctl_${ReleaseVersion}_windows_${Arch}.tar.gz.cert

# TODO: Add the checksum verification
# file=$(sha256sum ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz | awk '{print $1}')
# checksum=$(cat ksctl_${RELEASE_VERSION}_checksums.txt | grep ksctl_${RELEASE_VERSION}_${OS}_${ARCH}.tar.gz | awk '{print $1}')

# if [[ $file != $checksum ]]; then
#   echo -e "${Red}Checksum didn't matched!${NoColor}"
#   exit 1
# else
#   echo -e "${Green}CheckSum are verified${NoColor}"
# fi

tar -xvf ksctl_${ReleaseVersion}_windows_${Arch}.tar.gz


$localAppDataPath = $env:LOCALAPPDATA
$ksctl = Join-Path "$localAppDataPath" 'ksctl'

Write-Information "Path of AppDataPath $ksctl"

New-Item -ItemType Directory -Force -Path $ksctl | Out-Null

Copy-Item ksctl.exe -Destination "$ksctl/" -Force | Out-Null

Remove-Item ksctl*


Write-Host "[V] Finished Installation" -ForegroundColor DarkGreen
Write-Host ""
Write-Host "To run ksctl globally, please follow these steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "    1. Run the following command as administrator: ``setx PATH `"`$env:path;$ksctl`" -m``"
