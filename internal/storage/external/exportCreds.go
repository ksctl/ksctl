package external

import (
	"fmt"

	"github.com/ksctl/ksctl/internal/storage/external/mongodb"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
)

func HandleCreds(store consts.KsctlStore) (map[string][]byte, error) {
	switch store {
	case consts.StoreLocal, consts.StoreK8s:
		return nil, fmt.Errorf("these are not external store")
	case consts.StoreExtMongo:
		return mongodb.ExportEndpoint()
	default:
		return nil, fmt.Errorf("Invalid store")
	}
}
