package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getCiliumComponentOverridings(p metadata.ComponentOverrides) (version *string, ciliumChartOverridings map[string]any) {
	ciliumChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "ciliumChartOverridings":
			if v, ok := v.(map[string]any); ok {
				ciliumChartOverridings = v
			}
		}
	}
	return
}

func setCiliumComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	ciliumChartOverridings map[string]any,
) {
	version = "latest"
	ciliumChartOverridings = map[string]any{}

	_version, _ciliumChartOverridings := getCiliumComponentOverridings(p)

	if _version != nil {
		version = *_version
	}

	if _ciliumChartOverridings != nil {
		ciliumChartOverridings = _ciliumChartOverridings
	} else {
		ciliumChartOverridings = nil
	}
	return
}

func CiliumStandardComponent(params metadata.ComponentOverrides) metadata.StackComponent {
	version, ciliumChartOverridings := setCiliumComponentOverridings(params)

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoName: "cilium",
			RepoUrl:  "https://helm.cilium.io/",
			Charts: []metadata.ChartOptions{
				{
					Name:            "cilium/cilium",
					Version:         version,
					ReleaseName:     "cilium",
					Namespace:       "kube-system",
					CreateNamespace: false,
					Args:            ciliumChartOverridings,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
