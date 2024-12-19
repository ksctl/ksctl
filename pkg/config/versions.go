package config

var (
	OCIVersion   = "main"
	OCIImgSuffix = ""
)

func GetOCIVersion() string {
	return OCIVersion
}

func GetOCIImgSuffix() string {
	return OCIImgSuffix
}
