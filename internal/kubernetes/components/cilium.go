package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
)

func CiliumStandardComponent(params metadata.ComponentParams) metadata.StackComponent {
	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoName: "cilium",
			RepoUrl:  "https://helm.cilium.io/",
			Charts: []metadata.HelmOptions{
				{
					ChartName:       "cilium/cilium",
					ChartVer:        params.Version,
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
