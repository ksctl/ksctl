package k3s

import (
	"fmt"
	"strings"
)

func isValidK3sVersion(ver string) error {
	validVersion := []string{"1.29.4", "1.27.4", "1.27.1", "1.26.7", "1.25.12"} // TODO: check

	for _, vver := range validVersion {
		if vver == ver {
			return nil
		}
	}
	return log.NewError(k3sCtx, "invalid k3s version", "valid versions", strings.Join(validVersion, " "))
}

func getEtcdMemberIPFieldForControlplane(ips []string) string {
	tempDS := []string{}
	for _, ip := range ips {
		newValue := fmt.Sprintf("https://%s:2379", ip)
		tempDS = append(tempDS, newValue)
	}

	return strings.Join(tempDS, ",")
}
