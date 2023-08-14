package civo

import (
	"errors"
	"github.com/civo/civogo"
	"testing"
)

var (
	fakeClient *civogo.FakeClient
)

func TestCivoProvider_InitState(t *testing.T) {

}

func TestCivoProvider_VMType(t *testing.T) {

}

func TestInitializeFakeClient(t *testing.T) {
	var err error
	fakeClient, err = civogo.NewFakeClient()
	if err != nil {
		t.Fatal("[civo_test] failed to initialize")
	}
}

func TestCivoProvider_Name(t *testing.T) {

}

func TestIsValidRegion(t *testing.T) {
	testSet := map[string]error{
		"LON1": nil,
		"FRA1": nil,
		"NYC1": nil,
		"Lon!": errors.New(""),
		"":     errors.New(""),
	}
	for region, expected := range testSet {
		// FIXME: want to use fakeClient but it uses real client
		if err := isValidRegion(region); !errors.Is(expected, err) {
			t.Fatalf("Region code mismatch %s\n", region)
		}
	}
}
