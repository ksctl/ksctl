package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ksctl/ksctl/pkg/types"
)

func GetReqPayload(l types.LoggerFactory) (Operation, types.Metadata) {
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

	var payload types.Metadata
	err = json.Unmarshal(raw, &payload)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	return Operation(*arg1), payload
}
