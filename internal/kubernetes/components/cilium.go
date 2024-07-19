package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getCiliumComponentOverridings(p metadata.ComponentOverriding) (version *string, ciliumChartOverridings map[string]any) {
	ciliumChartOverridings = nil // By default it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			version = utilities.Ptr(v.(string))
		case "ciliumChartOverridings":
			ciliumChartOverridings = v.(map[string]any)
		}
	}
	return
}

func CiliumStandardComponent(params metadata.ComponentOverriding) metadata.StackComponent {
	var (
		version                = "latest"
		ciliumChartOverridings = map[string]any{}
	)

	_version, _ciliumChartOverridings := getCiliumComponentOverridings(params)

	if _version != nil {
		version = *_version
	}

	if _ciliumChartOverridings != nil {
		ciliumChartOverridings = _ciliumChartOverridings
	} else {
		ciliumChartOverridings = nil
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoName: "cilium",
			RepoUrl:  "https://helm.cilium.io/",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "cilium/cilium",
					ChartVer:        version,
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
