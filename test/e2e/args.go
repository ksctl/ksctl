package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ksctl/ksctl/pkg/resources"
)

func GetReqPayload(l resources.LoggerFactory) (Operation, resources.Metadata) {
	l.SetPackageName("e2e-tests")
	arg1 := flag.String("op", "", "operation to perform")
	arg2 := flag.String("file", "", "file name as payload")

	// Parse the command-line flags
	flag.Parse()

	// Check if required arguments are provided
	if *arg1 == "" || *arg2 == "" {
		fmt.Println("Usage: go run log.go -op <value> -file <value>")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*arg2)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	var payload resources.Metadata
	err = json.Unmarshal(raw, &payload)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	verbosityLevel := 0
	if os.Getenv("E2E_LOG_LEVEL") == "DEBUG" {
		verbosityLevel = -1
	}

	payload.LogVerbosity = verbosityLevel
	payload.LogWritter = os.Stdout

	return Operation(*arg1), payload
}
