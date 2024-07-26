package metadata

type ChartOptions struct {
	Version         string
	Name            string
	ReleaseName     string
	Namespace       string
	CreateNamespace bool
	Args            map[string]interface{}
}

type KubectlHandler struct {
	CreateNamespace bool
	Namespace       string

	Version     string // how to get the version propogated?
	Url         string
	PostInstall string
	Metadata    string
}

type HelmHandler struct {
	RepoUrl  string
	RepoName string
	Charts   []ChartOptions
}

type (
	StackComponentType uint
	StackComponentID   string
	StackID            string
)

type StackComponent struct {
	Helm        *HelmHandler
	Kubectl     *KubectlHandler
	HandlerType StackComponentType
}

// ApplicationStack TODO(dipankar): need to think of taking some sport of the application ksctl provide from the src to some json file in ver control
//
//	so that we can update that and no need of update of the logicial part
//
// also add a String()
type ApplicationStack struct {
	Components map[StackComponentID]StackComponent

	// StkDepsIdx helps you to get sequence of components, aka it acts as a key value table
	StkDepsIdx []StackComponentID

	Maintainer  string
	StackNameID StackID
}

type ApplicationParams struct {
	// StkOverrides   map[string]any
	ComponentParams map[StackComponentID]ComponentOverrides
}

type ComponentOverrides map[string]any
