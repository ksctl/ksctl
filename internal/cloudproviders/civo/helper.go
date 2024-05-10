package civo

import (
	"os"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

// fetchAPIKey returns the api_token from the cred/civo.json file store
func fetchAPIKey(storage resources.StorageFactory) string {

	civoToken := os.Getenv("CIVO_TOKEN")
	if civoToken != "" {
		return civoToken
	}
	log.Warn("environment vars not set: `CIVO_TOKEN`")

	credentials, err := storage.ReadCredentials(consts.CloudCivo)
	if err != nil {
		return ""
	}
	if credentials.Civo == nil {
		return ""
	}
	return credentials.Civo.Token
}

func loadStateHelper(storage resources.StorageFactory) error {
	raw, err := storage.Read()
	if err != nil {
		return log.NewError(err.Error())
	}
	*mainStateDocument = func(x *types.StorageDocument) types.StorageDocument {
		return *x
	}(raw)
	return nil
}

// helper functions to get resources from civogo client
// seperation so that we can test logic by assert
func getValidK8sVersionClient(obj *CivoProvider) []string {
	vers, err := obj.client.ListAvailableKubernetesVersions()
	if err != nil {
		log.Error("unable to get available k8s versions", "err", err)
		return nil
	}
	log.Debug("Printing", "ListAvailableKubernetesVersions", vers)
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(consts.K8sK3s) {
			val = append(val, ver.Label)
		}
	}
	return val
}

func getValidRegionsClient(obj *CivoProvider) []string {
	regions, err := obj.client.ListRegions()
	if err != nil {
		log.Error("unable to get available regions", "err", err)
		return nil
	}
	log.Debug("Printing", "ListRegions", regions)
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val
}

func getValidVMSizesClient(obj *CivoProvider) []string {
	nodeSizes, err := obj.client.ListInstanceSizes()
	if err != nil {
		log.Error("unable to fetch list of valid instance sizes", "err", err)
		return nil
	}
	log.Debug("Printing", "ListInstanceSizes", nodeSizes)
	var val []string
	for _, region := range nodeSizes {
		val = append(val, region.Name)
	}
	return val
}

func validationOfArguments(obj *CivoProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *CivoProvider, ver string) error {
	var valver []string = getValidK8sVersionClient(obj)
	for _, vver := range valver {
		if vver == ver {
			return nil
		}
	}
	return log.NewError("Invalid k8s version\nValid options: %v\n", valver)
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegion(obj *CivoProvider, reg string) error {
	var validFromClient []string = getValidRegionsClient(obj)
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return log.NewError("INVALID REGION\nValid options: %v\n", validFromClient)
}

func isValidVMSize(obj *CivoProvider, size string) error {
	var validFromClient []string = getValidVMSizesClient(obj)
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return log.NewError("INVALID VM SIZE\nValid options: %v\n", validFromClient)
}
