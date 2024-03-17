package k3s

import (
	"fmt"
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

func getEtcdMemberIPFieldForDatastore(ips []string) string {
	tempDS := []string{}
	for idx, ip := range ips {
		newValue := fmt.Sprintf("infra%d=https://%s:2380", idx, ip)
		tempDS = append(tempDS, newValue)
	}

	return strings.Join(tempDS, ",")
}

func getEtcdMemberIPFieldForControlplane(ips []string) string {
	tempDS := []string{}
	for _, ip := range ips {
		newValue := fmt.Sprintf("https://%s:2379", ip)
		tempDS = append(tempDS, newValue)
	}

	return strings.Join(tempDS, ",")
}
