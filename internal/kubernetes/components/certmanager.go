package components

import (
	"slices"
	"strings"

	"github.com/ksctl/ksctl/poller"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getCertManagerComponentOverridings(p metadata.ComponentOverrides) (
	version *string,
	gateway_apiEnable *bool,
	certmanagerChartOverridings map[string]any,
) {
	if p == nil {
		return nil, nil, nil
	}
	certmanagerChartOverridings = nil

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "certmanagerChartOverridings":
			if v, ok := v.(map[string]any); ok {
				certmanagerChartOverridings = v
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
	err error,
) {

	releases, err := poller.GetSharedPoller().Get("cert-manager", "cert-manager")
	if err != nil {
		return "", nil, err
	}
	version = releases[0]

	overridings = map[string]any{
		"crds": map[string]any{
			"enabled": "true",
		},
	}

	_version, _gateway_apiEnable, _certmanagerChartOverridings := getCertManagerComponentOverridings(params)

	if _version != nil {
		version = *_version
	}

	version = strings.TrimPrefix(version, "v")

	if _certmanagerChartOverridings != nil {
		utilities.CopySrcToDestPreservingDestVals(overridings, _certmanagerChartOverridings)
	}

	if _gateway_apiEnable != nil { // TODO: need to see later on how
		if *_gateway_apiEnable {
			if v, ok := overridings["extraArgs"]; ok {
				if v, ok := v.([]string); ok {
					if ok := slices.Contains[[]string, string](v, "--enable-gateway-api"); !ok {
						overridings["extraArgs"] = append(v, "--enable-gateway-api")
					}
				}
			} else {
				overridings["extraArgs"] = []string{"--enable-gateway-api"}
			}
		}
	}

	return
}

func CertManagerComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {
	version, overridings, err := setCertManagerComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

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
	}, nil
}
