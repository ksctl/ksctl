package ha_civo

import (
	"fmt"
	"testing"
)

// TODO: ADD TEST CASES

func TestGenerateDBPassword(t *testing.T) {
	fmt.Println(generateDBPassword(20))
}
