package components

import (
	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

const kwasmOperatorChartOverridingsKey = "kwasmOperatorChartOverridings"

func getKwasmOperatorComponentOverridings(p metadata.ComponentOverrides) (
	version *string,
	kwasmOperatorChartOverridings map[string]any,
) {
	kwasmOperatorChartOverridings = nil // By default it is nil
	if p == nil {
		return nil, nil
	}

	if v, ok := p["version"]; ok {
		if v, ok := v.(string); ok {
			version = utilities.Ptr(v)
		}
	}

	if v, ok := p[kwasmOperatorChartOverridingsKey]; ok {
		if v, ok := v.(map[string]any); ok {
			kwasmOperatorChartOverridings = v
		}
	}

	return
}

func setKwasmOperatorComponentOverridings(params metadata.ComponentOverrides) (
	version string,
	overridings map[string]any,
	err error,
) {
	version = "latest"

	_version, _kwasmOperatorChartOverridings := getKwasmOperatorComponentOverridings(params)

	if _version != nil {
		version = *_version
	}

	if _kwasmOperatorChartOverridings != nil {
		overridings = utilities.DeepCopyMap(_kwasmOperatorChartOverridings)
	}

	return
}

func KwasmOperatorComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {
	version, kwasmOperatorChartOverridings, err := setKwasmOperatorComponentOverridings(params)
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return metadata.StackComponent{
		Helm: &metadata.HelmHandler{
			RepoName: "kwasm",
			RepoUrl:  "http://kwasm.sh/kwasm-operator/",
			Charts: []metadata.ChartOptions{
				{
					Name:            "kwasm/kwasm-operator",
					Version:         version,
					ReleaseName:     "kwasm-operator",
					Namespace:       "kwasm",
					CreateNamespace: true,
					Args:            kwasmOperatorChartOverridings,
				},
			},
		},
	}, nil
}
