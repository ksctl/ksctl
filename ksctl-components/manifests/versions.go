// Dont edit it unless you know what you are doing
// This file contains the version of the ksctl agent and ksctl state import
// used to provide the ksctl core which version to fetch based on ksctl branch
// KsctlApplicationStackBranchOrTagName CAUTION: make sure it is restored to `latest`
// Once the dev is done
package manifests

const (
	KsctlAgentAppVersion                 = "latest"
	KsctlStateImportAppVersion           = "latest"
	KsctlApplicationStackBranchOrTagName = "latest" // NOTE: it is the branch name
)
