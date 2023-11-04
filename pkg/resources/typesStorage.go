package resources

import (
	"os"
)

type StorageFactory interface {
	// Save the data in bytes to specific location
	Save([]byte) error

	// TODO: check if required
	Destroy() error

	// Load gets contenets of file in bytes
	Load() ([]byte, error)

	// Path setter for path
	Path(string) StorageFactory

	// Permission setter for permission
	Permission(mode os.FileMode) StorageFactory

	// CreateDir creates directory
	CreateDir() error

	// DeleteDir deletes directories
	DeleteDir() error

	// GetFolders returns the folder's contents
	GetFolders() ([][]string, error)
}
