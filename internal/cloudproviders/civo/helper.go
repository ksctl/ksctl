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
	log.Warn(civoCtx, "environment vars not set: `CIVO_TOKEN`")

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
		return err
	}
	*mainStateDocument = func(x *types.StorageDocument) types.StorageDocument {
		return *x
	}(raw)
	return nil
}

// helper functions to get resources from civogo client
// seperation so that we can test logic by assert
func getValidK8sVersionClient(obj *CivoProvider) ([]string, error) {
	vers, err := obj.client.ListAvailableKubernetesVersions()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListAvailableKubernetesVersions", vers)
	var val []string
	for _, ver := range vers {
		if ver.ClusterType == string(consts.K8sK3s) {
			val = append(val, ver.Label)
		}
	}
	return val, nil
}

func getValidRegionsClient(obj *CivoProvider) ([]string, error) {
	regions, err := obj.client.ListRegions()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListRegions", regions)
	var val []string
	for _, region := range regions {
		val = append(val, region.Code)
	}
	return val, nil
}

func getValidVMSizesClient(obj *CivoProvider) ([]string, error) {
	nodeSizes, err := obj.client.ListInstanceSizes()
	if err != nil {
		return nil, err
	}
	log.Debug(civoCtx, "Printing", "ListInstanceSizes", nodeSizes)
	var val []string
	for _, region := range nodeSizes {
		val = append(val, region.Name)
	}
	return val, nil
}

func validationOfArguments(obj *CivoProvider) error {

	if err := isValidRegion(obj, obj.region); err != nil {
		return err
	}

	return nil
}

func isValidK8sVersion(obj *CivoProvider, ver string) error {
	valver, err := getValidK8sVersionClient(obj)
	if err != nil {
		return err
	}
	for _, vver := range valver {
		if vver == ver {
			return nil
		}
	}
	return log.NewError(civoCtx, "invalid k8s version", "Valid options", valver)
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegion(obj *CivoProvider, reg string) error {
	validFromClient, err := getValidRegionsClient(obj)
	if err != nil {
		return err
	}
	for _, region := range validFromClient {
		if region == reg {
			return nil
		}
	}
	return log.NewError(civoCtx, "invalid region", "Valid options", validFromClient)
}

func isValidVMSize(obj *CivoProvider, size string) error {
	validFromClient, err := getValidVMSizesClient(obj)
	if err != nil {
		return err
	}
	for _, nodeSize := range validFromClient {
		if size == nodeSize {
			return nil
		}
	}
	return log.NewError(civoCtx, "invalid VM size", "Valid options", validFromClient)
}
