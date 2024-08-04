package kubernetes

import "testing"

func TestAppManifestFetch(t *testing.T) {
	resUrl, err := getManifests("https://raw.githubusercontent.com/ksctl/ksctl/main/ksctl-components/manifests/controllers/application/deploy.yml")
	if err != nil {
		t.Fatalf("getManifests failed: %s", err.Error())
	}
	if len(resUrl) == 0 {
		t.Fatalf("getManifests failed: %s", "no manifest found")
	}
}
