package components

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/kubernetes/metadata"
	"github.com/ksctl/ksctl/pkg/helpers/utilities"
)

func getSpinkubeComponentOverridings(p metadata.ComponentOverrides) (version *string) {
	if p == nil {
		return nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		}
	}
	return
}

func setSpinkubeComponentOverridings(p metadata.ComponentOverrides, theThing string) (
	version string,
	url string,
	postInstall string,
	err error,
) {

	version, err = utilities.GetLatestRepoRelease("spinkube", "spin-operator")
	if err != nil {
		return
	}

	url = ""
	postInstall = ""

	_version := getSpinkubeComponentOverridings(p)
	if _version != nil {
		version = *_version
	}

	defaultVals := func() {
		url = fmt.Sprintf("https://github.com/spinkube/spin-operator/releases/download/%s/%s", version, theThing)
		postInstall = "https://www.spinkube.dev/docs/topics/"
	}

	defaultVals()
	return
}

func SpinkubeOperatorCrdComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.crds.yaml")
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func SpinkubeOperatorRuntimeClassComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.runtime-class.yaml")
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func SpinkubeOperatorShimExecComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, url, postInstall, err := setSpinkubeComponentOverridings(params, "spin-operator.shim-executor.yaml")
	if err != nil {
		return metadata.StackComponent{}, err
	}

	return spinkubeReturnHelper(version, url, postInstall)
}

func spinkubeReturnHelper(version, url, postInstall string) (metadata.StackComponent, error) {
	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeKubectl,
		Kubectl: &metadata.KubectlHandler{
			Url:             url,
			Version:         version,
			CreateNamespace: false,
			Metadata:        fmt.Sprintf("KubeSpin (ver: %s) is an open source project that streamlines developing, deploying and operating WebAssembly workloads in Kubernetes - resulting in delivering smaller, more portable applications and incredible compute performance benefits", version),
			PostInstall:     postInstall,
		},
	}, nil
}

func SpinOperatorComponent(params metadata.ComponentOverrides) (metadata.StackComponent, error) {

	version, helmOverride := setSpinOperatorComponentOverridings(params)

	return metadata.StackComponent{
		HandlerType: metadata.ComponentTypeHelm,
		Helm: &metadata.HelmHandler{
			Charts: []metadata.ChartOptions{
				{
					Name:            "spinkube/spin-operator",
					Version:         version,
					ReleaseName:     "spin-operator",
					Namespace:       "spin-operator",
					CreateNamespace: true,
					Args:            helmOverride,
					ChartRef:        "oci://ghcr.io/spinkube/charts/spin-operator",
				},
			},
		},
	}, nil
}

func getSpinkubeOperatorComponentOverridings(p metadata.ComponentOverrides) (version *string, helmOperatorChartOverridings map[string]interface{}) {
	helmOperatorChartOverridings = nil // By default, it is nil

	if p == nil {
		return nil, nil
	}

	for k, v := range p {
		switch k {
		case "version":
			if v, ok := v.(string); ok {
				version = utilities.Ptr(v)
			}
		case "helmOperatorChartOverridings":
			if v, ok := v.(map[string]interface{}); ok {
				helmOperatorChartOverridings = v
			}
		}
	}
	return
}

func setSpinOperatorComponentOverridings(p metadata.ComponentOverrides) (
	version string,
	helmOperatorChartOverridings map[string]any,
) {
	version = "latest"
	helmOperatorChartOverridings = map[string]any{}

	_version, _helmOperatorChartOverridings := getSpinkubeOperatorComponentOverridings(p)

	if _version != nil {
		version = *_version
	}

	if _helmOperatorChartOverridings != nil {
		helmOperatorChartOverridings = _helmOperatorChartOverridings
	}

	return
}
