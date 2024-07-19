package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getCiliumComponentOverridings(p metadata.ComponentOverriding) (version *string) {
	if p == nil {
		return nil
	}

	for k, v := range p {
		switch k {
		case "version":
			version = utilities.Ptr(v.(string))
		}
	}
	return
}

func CiliumStandardComponent(params metadata.ComponentOverriding) metadata.StackComponent {
	var (
		version = "latest"
	)
	_version := getCiliumComponentOverridings(params)
	if _version != nil {
		version = *_version
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
					Args:            nil,
				},
			},
		},
		HandlerType: metadata.ComponentTypeHelm,
	}
}
