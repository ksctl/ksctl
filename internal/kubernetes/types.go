package kubernetes

type InstallType string

type Application struct {
	Name          string
	Url           string
	Version       string
	Maintainer    string
	HelmConfig    []HelmOptions
	KubectlConfig KubectlOptions
	InstallType
}

type KubectlOptions struct {
	createNamespace bool
	postInstall     string
	metadata        string
	// Namespace Only specify if createNamespace is true
	namespace string
}

type HelmOptions struct {
	chartVer        string
	chartName       string
	releaseName     string
	namespace       string
	createNamespace bool
	args            map[string]interface{}
}

type HelmHandler struct {
	createNamespace bool
	namespace       string

	postInstall string
	metadata    string
}

type KubectlHandler struct {
	namespace       string
	createNamespace bool

	chartVer    string
	chartName   string
	releaseName string
	args        map[string]interface{}
}

type StackComponent struct {
	helm    *HelmHandler
	kubectl *KubectlHandler
}

type ApplicationStack struct {
	// components it marks for the sequqncial dependency for each component
	components []StackComponent
}
