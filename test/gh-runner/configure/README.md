## Things which are needed to be configured in github custom-runners

### Os Level configurations

1. Create a new User `runner`
```bash
sudo usermod -m -s /bin/bash runner
```
2. Update and install the new updates
```bash
sudo apt update -y
```
3. install jq, tree, kubectl, golang, docker
```bash
sudo apt install jq tree -y

curl --fail-with-body -sSL https://get.docker.io | bash
```

go
```bash
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz

#### Make sure the path /usr/local/go/bin is added to the .path of the github runner
```
Kubectl
```bash
sudo apt-get update
# apt-transport-https may be a dummy package; if so, you can skip that package
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg

# If the folder `/etc/apt/keyrings` does not exist, it should be created before the curl command, read the note below.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl --fail-with-body -sSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg --yes
sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg # allow unprivileged APT programs to read this keyring

# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list   # helps tools such as command-not-found to work correctly

sudo apt-get update
sudo apt-get install -y kubectl
```

### Github specific configurations

1. Add it to ksctl custom runners list
> [!NOTE]
> the script changes from time to time
```bash
# Create a folder
mkdir actions-runner && cd actions-runner

# Download the latest runner package
curl --fail-with-body -o actions-runner-linux-x64-2.317.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.317.0/actions-runner-linux-x64-2.317.0.tar.gz

# Optional: Validate the hash
echo "9e883d210df8c6028aff475475a457d380353f9d01877d51cc01a17b2a91161d  actions-runner-linux-x64-2.317.0.tar.gz" | shasum -a 256 -c

# Extract the installer
tar xzf ./actions-runner-linux-x64-2.317.0.tar.gz
```

2. Create a directory `actions-runner/ksctl-bin` to hold the binary of ksctl

3. Once the directory for github custom runner setup is done make sure the ~/action-runner/.path has these entries as well
```conf
/usr/loca........:/usr/local/go/bin:/home/runner/actions-runner/ksctl-bin
```

https://docs.github.com/en/rest/actions/self-hosted-runners?apiVersion=2022-11-28#create-a-registration-token-for-a-repository
4. start or removal of github service
```bash
./config.sh --url https://github.com/ksctl/ksctl --token XXYYZZ --unattended --labels e2e
sudo ./svc.sh install
sudo ./svc.sh start
```

for removal
```bash
# need to see how this token behaves
./config.sh remove --url https://github.com/ksctl/ksctl --token XXYYZZ  --unattended
```


terraform uses this for public ip

terraform output -json ip_address | jq -r .

so for getting the results is this
```
[
  "2a01:XXBBB:dwecdfe:fff::1",
  "2a01:XXBBB:dwecdfe:csdcds::1"
]
```
