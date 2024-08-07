package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getCertManagerComponentOverridings(p metadata.ComponentOverrides) (version *string, gateway_apiEnable *bool) {
	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "gatewayapiEnable":
			if v, ok := v.(bool); ok {
				gateway_apiEnable = utilities.Ptr(v)
			}
		}
	}
	return
}

func setCertManagerComponentOverridings(params metadata.ComponentOverrides) (
	version string,
	overridings map[string]any,
) {
	version = "latest"
	overridings = map[string]any{
		"crds.enabled": true,
	}

	_version, _gateway_apiEnable := getCertManagerComponentOverridings(params)

	if _version != nil {
		version = *_version
	}

	if _gateway_apiEnable != nil {
		if *_gateway_apiEnable {
			overridings["extraArgs"] = []string{
				"--enable-gateway-api",
			}
		}
	}

	return
}

func CertManagerComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, overridings := setCertManagerComponentOverridings(params)

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeHelm,
		Helm: &metadata.HelmHandler{
			RepoUrl:  "https://charts.jetstack.io",
			RepoName: "jetstack",
			Charts: []metadata.ChartOptions{
				{
					Name:            "jetstack/cert-manager",
					Version:         version,
					ReleaseName:     "cert-manager",
					CreateNamespace: true,
					Namespace:       "cert-manager",
					Args:            overridings,
				},
			},
		},
	}
}
