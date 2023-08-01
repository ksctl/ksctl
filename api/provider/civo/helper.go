package civo

import (
	"fmt"
	"strings"

	"github.com/kubesimplify/ksctl/api/provider/utils"
)

func validationOfArguments(name, region string) error {

	if !isValidRegionCIVO(region) {
		return fmt.Errorf("REGION")
	}

	if !utils.IsValidName(name) {
		return fmt.Errorf("NAME FORMAT")
	}

	return nil
}

// IsValidRegionCIVO validates the region code for CIVO
func isValidRegionCIVO(reg string) bool {
	return strings.Compare(reg, "FRA1") == 0 ||
		strings.Compare(reg, "NYC1") == 0 ||
		strings.Compare(reg, "PHX1") == 0 ||
		strings.Compare(reg, "LON1") == 0
}
