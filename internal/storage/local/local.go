package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/ksctl/ksctl/pkg/logger"

	"github.com/goccy/go-json"

	"github.com/ksctl/ksctl/internal/storage/types"
	"github.com/ksctl/ksctl/pkg/helpers"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/resources"
)

const (
	filePerm       = os.FileMode(0640)
	dirPerm        = os.FileMode(0750)
	credentialPerm = os.FileMode(0600)
	subDirState    = "state"
	subDirCreds    = "credentials"
)

var log resources.LoggerFactory

type Store struct {
	cloudProvider string
	clusterType   string
	clusterName   string
	region        string
	userid        any
	mu            *sync.RWMutex
	wg            *sync.WaitGroup
}

func InitStorage(logVerbosity int, logWriter io.Writer) resources.StorageFactory {
	log = logger.NewDefaultLogger(logVerbosity, logWriter)
	log.SetPackageName(string(consts.StoreLocal))

	return &Store{mu: &sync.RWMutex{}, wg: &sync.WaitGroup{}}
}

func (db *Store) disconnect() error {
	return nil
}

func (db *Store) Kill() error {
	db.wg.Wait()
	defer log.Success("Local Storage Got Killed")

	return db.disconnect()
}

func (db *Store) Connect(ctx context.Context) error {
	db.userid = ctx.Value("USERID")

	log.Success("CONN to HostOS")
	return nil
}

func genOsClusterPath(creds bool, subDir ...string) string {

	var userLoc string
	if v := os.Getenv(string(consts.KsctlCustomDirEnabled)); len(v) != 0 {
		userLoc = strings.Join(strings.Split(strings.TrimSpace(v), " "), helpers.PathSeparator)
	} else {
		userLoc = helpers.GetUserName()
	}
	subKsctlLoc := subDirState
	if creds {
		subKsctlLoc = subDirCreds
	}
	pathArr := []string{userLoc, ".ksctl", subKsctlLoc}

	if !creds {
		pathArr = append(pathArr, subDir...)
	}
	log.Debug("storage.local.genOsClusterPath", "userLoc", userLoc, "subKsctlLoc", subKsctlLoc, "pathArr", pathArr)
	return strings.Join(pathArr, helpers.PathSeparator)
}

func reader(loc string) (*types.StorageDocument, error) {
	data, err := os.ReadFile(loc)
	if err != nil {
		return nil, err
	}

	var v *types.StorageDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (db *Store) Read() (*types.StorageDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if e := db.clusterPresent(nil); e != nil {
		return nil, e
	}
	dirPath := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region, "state.json")
	log.Debug("storage.local.Read", "dirPath", dirPath)
	return reader(dirPath)
}

func (db *Store) Write(v *types.StorageDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	dirPath := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)
	FileLoc := ""
	log.Debug("storage.local.Write", "dirPath", dirPath)

	if err := db.clusterPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region), dirPerm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cluster present: %w", err)
	}

	FileLoc = dirPath + helpers.PathSeparator + "state.json"
	log.Debug("storage.local.Write", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {
		// if the error occurs cleanup the directry
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("unable to cleanup after failure: %w", err)

		}
		return err
	}
	return os.WriteFile(FileLoc, data, filePerm)
}

func (db *Store) ReadCredentials(cloud consts.KsctlCloud) (*types.CredentialsDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	// for now create multiple files civo.json, azure.json, etc.
	if e := db.credentialsPresent(nil); e != nil {
		return nil, e
	}
	dirPath := genOsClusterPath(true)

	data, err := os.ReadFile(dirPath + helpers.PathSeparator + string(cloud) + ".json")
	if err != nil {
		return nil, err
	}

	var v *types.CredentialsDocument
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return v, nil
}

func (db *Store) WriteCredentials(cloud consts.KsctlCloud, v *types.CredentialsDocument) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.wg.Add(1)
	defer db.wg.Done()

	if err := db.credentialsPresent(func(err error) error {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(genOsClusterPath(true), dirPerm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("cluster present: %w", err)
	}

	dirPath := genOsClusterPath(true)
	FileLoc := ""

	FileLoc = dirPath + helpers.PathSeparator + string(cloud) + ".json"
	log.Debug("storage.local.WriteCredentials", "FileLoc", FileLoc)

	data, err := json.Marshal(v)
	if err != nil {

		return err
	}
	return os.WriteFile(FileLoc, data, credentialPerm)
}

func (db *Store) Setup(cloud consts.KsctlCloud, region, clusterName string, clusterType consts.KsctlClusterType) error {
	switch cloud {
	case consts.CloudAws, consts.CloudAzure, consts.CloudCivo, consts.CloudLocal:
		db.cloudProvider = string(cloud)
	default:
		return errors.New("invalid cloud")
	}
	if clusterType != consts.ClusterTypeHa && clusterType != consts.ClusterTypeMang {
		return errors.New("invalid cluster type")
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
		return fmt.Errorf("cluster not present: %w", e)
	}
	dirPath := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)

	return os.RemoveAll(dirPath)
}

// clusterPresent ir returns error when we have to create the folder when that folder is not present
// and if its present it returns nil
func (db *Store) clusterPresent(f func(error) error) error {
	dirPath := genOsClusterPath(false, db.cloudProvider, db.clusterType, db.clusterName+" "+db.region)
	_, err := os.Stat(dirPath)
	if err != nil {
		log.Debug("storage.local.clusterPresent", "err", err)
		if f != nil {
			return f(err)
		}
		return err
	}

	return nil
}

func (db *Store) credentialsPresent(f func(error) error) error {
	dirPath := genOsClusterPath(true)
	_, err := os.Stat(dirPath)
	if err != nil {
		if f != nil {
			if e := f(err); e != nil {
				return e
			}
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

	return db.clusterPresent(nil)
}

func (db *Store) GetOneOrMoreClusters(filter map[string]string) (map[consts.KsctlClusterType][]*types.StorageDocument, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	db.wg.Add(1)
	defer db.wg.Done()

	clusterType := filter["clusterType"]
	cloud := filter["cloud"]

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
	log.Debug("storage.local.GetOneOrMoreClusters", "filter", filter, "filterCloudPath", filterCloudPath, "filterClusterType", filterClusterType)

	clustersInfo := make(map[consts.KsctlClusterType][]*types.StorageDocument)

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
			log.Debug("storage.local.GetOneOrMoreClusters", "clusterInfo", clustersInfo)
		}
	}

	return clustersInfo, nil
}

func getClustersInfo(locs []string) ([]*types.StorageDocument, error) {
	var data []*types.StorageDocument

	for _, loc := range locs {
		v, err := reader(loc)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		data = append(data, v)
	}
	log.Debug("storage.local.getClustersInfo", "data", data)

	return data, nil
}

func fetchFilePaths(cloud string, clusterType string) ([]string, error) {
	dirPath := genOsClusterPath(false, cloud, clusterType)

	folders, err := os.ReadDir(dirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	log.Debug("storage.local.fetchFilePaths", "folders", folders)
	var info []string
	for _, file := range folders {
		if file.IsDir() {
			info = append(info, dirPath+helpers.PathSeparator+file.Name()+helpers.PathSeparator+"state.json")
		}
	}

	return info, nil
}
