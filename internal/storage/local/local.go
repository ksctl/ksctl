package local

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"
	"sync"

	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"

	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

const (
	filePerm       = os.FileMode(0640)
	dirPerm        = os.FileMode(0750)
	credentialPerm = os.FileMode(0600)
	subDirState    = "state"
	subDirCreds    = "credentials"
)

var (
	log      types.LoggerFactory
	storeCtx context.Context
)

type Store struct {
	cloudProvider string
	clusterType   string
	clusterName   string
	region        string
	userid        any
	mu            *sync.RWMutex
	wg            *sync.WaitGroup
}

func copyStore(src *Store, dest *Store) {
	dest.cloudProvider = src.cloudProvider
	dest.clusterName = src.clusterName
	dest.clusterType = src.clusterType
	dest.region = src.region
	dest.userid = src.userid
}

func (s *Store) PresentDirectory(_path []string) (loc string, isPresent bool) {
	loc = path.Join(_path...)
	_, err := os.ReadDir(loc)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	isPresent = true
	return
}

func (s *Store) CreateFileIfNotPresent(_path []string) (loc string, err error) {
	loc = path.Join(_path...)

	if _, err = os.ReadFile(loc); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if _, _err := os.Create(loc); _err != nil {
				err = ksctlErrors.ErrInternal.Wrap(
					log.NewError(storeCtx, "unable to create file", "Reason", _err),
				)
				return
			}
			err = nil
			return
		}
		err = ksctlErrors.ErrUnknown.Wrap(
			log.NewError(storeCtx, "failed to read the file", "Reason", err),
		)
		return
	}
	return
}

func (s *Store) CreateDirectory(_path []string) error {
	loc := path.Join(_path...)

	if err := os.MkdirAll(loc, dirPerm); err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to mkdir all folders", "Reason", err),
		)

	}
	return nil
}

func (s *Store) Export(filters map[consts.KsctlSearchFilter]string) (*types.StorageStateExportImport, error) {

	var cpyS *Store = s
	copyStore(s, cpyS) // for storing the state of the store before import was called!

	dest := new(types.StorageStateExportImport)

	_cloud := filters[consts.Cloud]
	_clusterType := filters[consts.ClusterType]
	_clusterName := filters[consts.Name]
	_region := filters[consts.Region]

	stateClustersForTypes, err := s.GetOneOrMoreClusters(map[consts.KsctlSearchFilter]string{
		consts.Cloud:       _cloud,
		consts.ClusterType: _clusterType,
	})
	if err != nil {
		return nil, err
	}

	for _, states := range stateClustersForTypes {
		// NOTE: make sure both filters are available if not then it will not apply
		if len(_clusterName) == 0 || len(_region) == 0 {
			dest.Clusters = append(dest.Clusters, states...)
		}
		for _, state := range states {
			if _clusterName == state.ClusterName &&
				_region == state.Region {
				dest.Clusters = append(dest.Clusters, state)
			}
		}
	}

	if len(_cloud) == 0 {
		// all the cloud provider credentials
		for _, constsCloud := range []consts.KsctlCloud{
			consts.CloudAws,
			consts.CloudCivo,
			consts.CloudLocal,
			consts.CloudAzure,
		} {
			_v, _err := s.ReadCredentials(constsCloud)

			if _err != nil {
				log.Debug(storeCtx, "error in reading credentials", "Reason", _err)
				if ksctlErrors.ErrNoMatchingRecordsFound.Is(_err) {
					// if errors.Is(_err, os.ErrNotExist) {
					continue
				} else {
					return nil, _err
				}
			}
			dest.Credentials = append(dest.Credentials, _v)
		}
	} else {
		_v, _err := s.ReadCredentials(consts.KsctlCloud(_cloud))
		if _err != nil && !errors.Is(_err, os.ErrNotExist) {
			return nil, _err
		}
		dest.Credentials = append(dest.Credentials, _v)
	}

	copyStore(cpyS, s) // for restoring the state of the store before import was called!
	return dest, nil
}

func (s *Store) Import(src *types.StorageStateExportImport) error {
	creds := src.Credentials
	states := src.Clusters

	var cpyS *Store = s
	copyStore(s, cpyS) // for storing the state of the store before import was called!

	for _, state := range states {
		cloud := state.InfraProvider
		region := state.Region
		clusterName := state.ClusterName
		clusterType := consts.KsctlClusterType(state.ClusterType)

		if err := s.Setup(cloud, region, clusterName, clusterType); err != nil {
			return err
		}

		if err := s.Write(state); err != nil {
			return err
		}
	}

	for _, cred := range creds {
		cloud := cred.InfraProvider
		if err := s.WriteCredentials(cloud, cred); err != nil {
			return err
		}
	}

	copyStore(cpyS, s) // for restoring the state of the store before import was called!
	return nil
}

func NewClient(parentCtx context.Context, _log types.LoggerFactory) *Store {
	storeCtx = context.WithValue(parentCtx, consts.KsctlModuleNameKey, string(consts.StoreLocal))
	log = _log
	return &Store{mu: &sync.RWMutex{}, wg: &sync.WaitGroup{}}
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer log.Success(storeCtx, "Local Storage Got Killed")

	return db.disconnect()
}

func (db *Store) Connect() error {
	if v, ok := helpers.IsContextPresent(storeCtx, consts.KsctlContextUserID); ok {
		db.userid = v
	} else {
		db.userid = "default"
	}

	log.Success(storeCtx, "CONN to HostOS")
	return nil
}

func genOsClusterPath(creds bool, subDir ...string) (string, error) {

	var userLoc string
	if v, ok := helpers.IsContextPresent(storeCtx, consts.KsctlCustomDirLoc); ok {
		userLoc = path.Join(strings.Split(strings.TrimSpace(v), " ")...)
	} else {
		v, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userLoc = v
	}
	subKsctlLoc := subDirState
	if creds {
		subKsctlLoc = subDirCreds
	}
	pathArr := []string{userLoc, ".ksctl", subKsctlLoc}

	if !creds {
		pathArr = append(pathArr, subDir...)
	}
	log.Debug(storeCtx, "storage.local.genOsClusterPath", "userLoc", userLoc, "subKsctlLoc", subKsctlLoc, "pathArr", pathArr)

	return path.Join(pathArr...), nil
}

func reader(loc string) (*storageTypes.StorageDocument, error) {
	data, err := os.ReadFile(loc)
	if err != nil {
		return nil, err
	}

	var v *storageTypes.StorageDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (db *Store) Read() (*storageTypes.StorageDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if e := db.clusterPresent(nil); e != nil {
		return nil, e
	}
	dirPath, err := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region, "state.json")
	if err != nil {
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	log.Debug(storeCtx, "storage.local.Read", "dirPath", dirPath)
	if v, e := reader(dirPath); e != nil {
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to read in host", "Reason", err),
		)
	} else {
		return v, nil
	}
}

func (db *Store) Write(v *storageTypes.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	dirPath, err := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}
	FileLoc := ""
	log.Debug(storeCtx, "storage.local.Write", "dirPath", dirPath)

	if err := db.clusterPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dirPath, dirPerm); err != nil {
				return ksctlErrors.ErrInternal.Wrap(
					log.NewError(storeCtx, "failure in creating directories", "Reason", err),
				)
			}
		}
		return nil
	}); err != nil {
		if !ksctlErrors.ErrInternal.Is(err) {
			return ksctlErrors.ErrDuplicateRecords.Wrap(
				log.NewError(storeCtx, "cluster already present", "Reason", err),
			)
		} else {
			return err
		}
	}

	FileLoc = path.Join(dirPath, "state.json")
	log.Debug(storeCtx, "storage.local.Write", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		// if the error occurs cleanup the directry
		// if err := os.RemoveAll(dirPath); err != nil { // TODO: why are we removing the directory
		// 	return fmt.Errorf("unable to cleanup after failure: %w", err)
		// }
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to serialize state", "Reason", err),
		)
	}
	if err := os.WriteFile(FileLoc, data, filePerm); err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to write in host", "Reason", err),
		)
	}
	return nil
}

func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*storageTypes.CredentialsDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	// for now create multiple files civo.json, azure.json, etc.
	if e := db.credentialsPresent(nil); e != nil {
		return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(storeCtx, "credentials are absent", "Reason", e),
		)
	}
	dirPath, err := genOsClusterPath(true)
	if err != nil {
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	data, err := os.ReadFile(path.Join(dirPath, string(cloud)+".json"))
	if err != nil {
		return nil, ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(storeCtx, "failed to read a host file", "Reason", err),
		)
	}

	var v *storageTypes.CredentialsDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, ksctlErrors.ErrNilCredentials.Wrap(
			log.NewError(storeCtx, "failed to deserialize the credentials", "Reason", err),
		)
	}

	return v, nil
}

func (db *Store) WriteCredentials(cloud consts.KsctlCloud, v *storageTypes.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	dirPath, err := genOsClusterPath(true)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	if err := db.credentialsPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dirPath, dirPerm); err != nil {
				return ksctlErrors.ErrInternal.Wrap(
					log.NewError(storeCtx, "failure in creating directories", "Reason", err),
				)
			}
		}
		return nil
	}); err != nil {
		if !ksctlErrors.ErrInternal.Is(err) {
			return ksctlErrors.ErrDuplicateRecords.Wrap( // TODO: need to check if it is a duplicate record error or not?
				log.NewError(storeCtx, "credentials already present", "Reason", err),
			)
		} else {
			return err
		}
	}

	FileLoc := ""

	FileLoc = path.Join(dirPath, string(cloud)+".json")
	log.Debug(storeCtx, "storage.local.WriteCredentials", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to serialize the credentials", "Reason", err),
		)
	}
	if e := os.WriteFile(FileLoc, data, credentialPerm); e != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to write file in host", "Reason", err),
		)
	}
	return nil
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return ksctlErrors.ErrInvalidCloudProvider.Wrap(
			log.NewError(storeCtx, "invalid", "cloud", cloud),
		)
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return ksctlErrors.ErrInvalidClusterType.Wrap(
			log.NewError(storeCtx, "invalid", "clusterType", clusterType),
		)
	}
	db.clusterName = clusterName
	db.region = region
	db.clusterType = string(clusterType)
	return nil
}

func (db *Store) DeleteCluster() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if e := db.clusterPresent(nil); e != nil {
		return ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(storeCtx, "cluster not present", "Reason", e),
		)
	}
	dirPath, err := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)
	if err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to perform complete clenup some directories are left behind", "Reason", err),
		)
	}
	return nil
}

func (db *Store) clusterPresent(handleErrFunc func(error) error) error {
	dirPath, _ := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)
	_, err := os.Stat(dirPath)
	if err != nil {
		log.Debug(storeCtx, "storage.local.clusterPresent", "err", err)
		if handleErrFunc != nil {
			return handleErrFunc(err)
		}
		return err
	}

	return nil
}

func (db *Store) credentialsPresent(handleErrFunc func(error) error) error {
	dirPath, _ := genOsClusterPath(true)
	_, err := os.Stat(dirPath)
	if err != nil {
		log.Debug(storeCtx, "storage.local.credentialsPresent", "err", err)
		if handleErrFunc != nil {
			return handleErrFunc(err)
		}
		return err
	}

	return nil
}

func (db *Store) AlreadyCreated(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	err := db.Setup(cloud, region, clusterName, clusterType)
	if err != nil {
		return err
	}

	if e := db.clusterPresent(nil); e != nil {
		return ksctlErrors.ErrNoMatchingRecordsFound.Wrap(
			log.NewError(storeCtx, "cluster not present", "Reason", e),
		)
	}
	return nil
}

func (db *Store) GetOneOrMoreClusters(filter map[consts.KsctlSearchFilter]string) (map[consts.KsctlClusterType][]*storageTypes.StorageDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	clusterType := filter[consts.ClusterType]
	cloud := filter[consts.Cloud]

	var filterCloudPath, filterClusterType []string

	switch cloud {
	case string(consts.CloudAll), "":
		filterCloudPath = append(filterCloudPath, string(consts.CloudCivo), string(consts.CloudAws), string(consts.CloudAzure), string(consts.CloudLocal))

	case string(consts.CloudCivo):
		filterCloudPath = append(filterCloudPath, string(consts.CloudCivo))

	case string(consts.CloudAzure):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAzure))

	case string(consts.CloudAws):
		filterCloudPath = append(filterCloudPath, string(consts.CloudAws))

	case string(consts.CloudLocal):
		filterCloudPath = append(filterCloudPath, string(consts.CloudLocal))
	}

	switch clusterType {
	case string(consts.ClusterTypeHa):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeHa))

	case string(consts.ClusterTypeMang):
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang))

	case "":
		filterClusterType = append(filterClusterType, string(consts.ClusterTypeMang), string(consts.ClusterTypeHa))
	}
	log.Debug(storeCtx, "storage.local.GetOneOrMoreClusters", "filter", filter, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*storageTypes.StorageDocument)

	for _, cloud := range filterCloudPath {
		for _, clusterType := range filterClusterType {
			files, err := fetchFilePaths(cloud, clusterType)
			if err != nil {
				return nil, err
			}
			v, err := getClustersInfo(files)
			if err != nil {
				return nil, err
			}

			clustersInfo[consts.KsctlClusterType(clusterType)] = append(clustersInfo[consts.KsctlClusterType(clusterType)], v...)
			log.Debug(storeCtx, "storage.local.GetOneOrMoreClusters", "clusterInfo", clustersInfo)
		}
	}

	return clustersInfo, nil
}

func getClustersInfo(locs []string) ([]*storageTypes.StorageDocument, error) {
	var data []*storageTypes.StorageDocument

	for _, loc := range locs {
		v, err := reader(loc)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, ksctlErrors.ErrInternal.Wrap(
				log.NewError(storeCtx, "failed to read in host", "Reason", err),
			)
		}
		data = append(data, v)
	}
	log.Debug(storeCtx, "storage.local.getClustersInfo", "data", data)

	return data, nil
}

func fetchFilePaths(cloud string, clusterType string) ([]string, error) {
	dirPath, err := genOsClusterPath(false, cloud, clusterType)
	if err != nil {
		return nil, ksctlErrors.ErrInternal.Wrap(
			log.NewError(storeCtx, "failed to gen clusterpath in host", "Reason", err),
		)
	}

	folders, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err // as the other function uses the err code getClustersInfo
	}

	log.Debug(storeCtx, "storage.local.fetchFilePaths", "folders", folders)
	var info []string
	for _, file := range folders {
		if file.IsDir() {
			info = append(info, path.Join(dirPath, file.Name(), "state.json"))
		}
	}

	return info, nil
}
