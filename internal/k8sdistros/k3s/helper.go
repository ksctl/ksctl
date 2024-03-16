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
	//infra0=https://192.168.1.2:2380,infra1=https://192.168.1.3:2380,infra2=https://192.168.1.4:2380
	fmt.Println("DS IPS:", ips)
	tempDS := []string{}
	for idx, ip := range ips {
		newValue := fmt.Sprintf("infra%d=https://%s:2380", idx, ip)
		tempDS = append(tempDS, newValue)
	}
	fmt.Println("Res:", tempDS)

	return strings.Join(tempDS, ",")
}

func getEtcdMemberIPFieldForControlplane(ips []string) string {
	//https://192.168.1.2:2379,https://192.168.1.3:2379,https://192.168.1.4:2379
	fmt.Println("CP IPS:", ips)
	tempDS := []string{}
	for _, ip := range ips {
		newValue := fmt.Sprintf("https://%s:2379", ip)
		tempDS = append(tempDS, newValue)
	}
	fmt.Println("Res:", tempDS)

	return strings.Join(tempDS, ",")
}
