package restapi

type RestApiServer interface {

	// it will send back the json as response

	Create(req interface{}) (res string, err error)
	Delete(req interface{}) (res string, err error)
	Switch(req interface{}) (res string, err error)
}
