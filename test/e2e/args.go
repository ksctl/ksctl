package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kubesimplify/ksctl/pkg/resources"
)

func GetReqPayload(l *log.Logger) (Operation, resources.Metadata) {

	arg1 := flag.String("op", "", "operation to perform")
	arg2 := flag.String("file", "", "file name as payload")

	// Parse the command-line flags
	flag.Parse()

	// Check if required arguments are provided
	if *arg1 == "" || *arg2 == "" {
		fmt.Println("Usage: go run main.go -op <value> -file <value>")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*arg2)
	if err != nil {
		l.Fatal(err)
	}

	var payload resources.Metadata
	err = json.Unmarshal(raw, &payload)
	if err != nil {
		l.Fatal(err)
	}

	return Operation(*arg1), payload
}
