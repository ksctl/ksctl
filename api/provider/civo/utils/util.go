package utils

import (
	"fmt"
	"strings"
)

// civo specific things

var (
	HOME_DIR   = "/home/dumy"                                    // it should be for all the providers (clouds)
	VALID_PATH = fmt.Sprintf("%s/.ksctl/config/civo/", HOME_DIR) // decide where it should be present
)

func fetchAPIKey() string {
	// it will be using the state manager to fetch the info
	return ""
}

func IsValidRegion(region string) bool {
	return strings.Compare(region, "FRA1") == 0 ||
		strings.Compare(region, "NYC1") == 0 ||
		strings.Compare(region, "PHX1") == 0 ||
		strings.Compare(region, "LON1") == 0
}
