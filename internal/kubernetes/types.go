package kubernetes

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
	components []StackComponent

	Maintainer string

	StackNameID string
}

type (
	applicationParams struct {
		version            string
		namespaceLvlAccess bool
		noUI               bool
		// namespace          string
	}
)
