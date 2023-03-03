package azure

import "testing"

func TestValidRegions(t *testing.T) {
	testData := []string{"abcd", "eastus", "westus2"}
	expectedResult := []bool{false, true, true}
	for i := 0; i < len(testData); i++ {
		if isValidRegion(testData[i]) != expectedResult[i] {
			t.Fatalf("%s region got %v but was expecting %v", testData[i], isValidRegion(testData[i]), expectedResult[i])
		}
	}
}

func TestValidNodeSizes(t *testing.T) {
	testData := []string{"Standard_D16d_v5", "Standard_F1", "Standard_sdcdd"}
	expectedResult := []bool{true, true, false}
	for i := 0; i < len(testData); i++ {
		if isValidNodeSize(testData[i]) != expectedResult[i] {
			t.Fatalf("%s region got %v but was expecting %v", testData[i], isValidNodeSize(testData[i]), expectedResult[i])
		}
	}
}

// TODO: Add more
