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

type KubectlHandler struct {
	createNamespace bool
	namespace       string

	version     string // how to get the version propogated?
	url         string
	postInstall string
	metadata    string
}

type HelmHandler struct {
	repoUrl  string
	repoName string
	charts   []HelmOptions
}

type (
	StackComponentType uint
)

const (
	ComponentTypeHelm    StackComponentType = iota
	ComponentTypeKubectl StackComponentType = iota
)

type StackComponent struct {
	helm        *HelmHandler
	kubectl     *KubectlHandler
	handlerType StackComponentType
}

// ApplicationStack TODO(dipankar): need to think about version peneatration from the stack later to the crd level
//
//	Also need to think of taking some sport of the application ksctl provide from the src to some json file in ver control
//	so that we can update that and no need of update of the logicial part
type ApplicationStack struct {
	// components it marks for the sequential dependency for each component
	components []StackComponent

	// StackType NOTE(dipankar):
	//		if `prod` components are treated as a part of the entire stack
	// Suitable for Production mode
	//		else `std` components are treated as loosly grouped or accomulated standard installation without keeping in mind whether the application are of similar nature or deps
	// Suitable for Development mode
	Maintainer string

	StackNameID string
}
