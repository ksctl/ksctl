package k3s

import (
	"strings"
)

func isValidK3sVersion(ver string) bool {
	validVersion := []string{"1.27.4", "1.27.1", "1.26.7", "1.25.12"} // TODO: check

	for _, vver := range validVersion {
		if vver == ver {
			return true
		}
	}
	log.Error(strings.Join(validVersion, " "))
	return false
}
