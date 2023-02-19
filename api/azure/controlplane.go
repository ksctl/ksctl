package azure

import "fmt"

func scriptWithoutCP_1(dbEndpoint, privateIPlb string) string {

	return fmt.Sprintf(`#!/bin/bash
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
`, dbEndpoint, privateIPlb)
}

func scriptWithCP_1() string {
	return `#!/bin/bash
cat /var/lib/rancher/k3s/server/token
`
}

func scriptCP_n(dbEndpoint, privateIPlb, token string) string {
	return fmt.Sprintf(`#!/bin/bash
export SECRET='%s'
export K3S_DATASTORE_ENDPOINT='%s'
curl -sfL https://get.k3s.io | sh -s - server \
	--token=$SECRET \
	--node-taint CriticalAddonsOnly=true:NoExecute \
	--tls-san %s
`, token, dbEndpoint, privateIPlb)
}

func scriptKUBECONFIG() string {
	return `#!/bin/bash
cat /etc/rancher/k3s/k3s.yaml`
}
