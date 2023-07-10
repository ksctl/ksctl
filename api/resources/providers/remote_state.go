
package providers

type RemoteStorage interface {
    Save(path string, data interface{}) error
    Load(path string) (interface{}, error) // try to make the return type defined
}
