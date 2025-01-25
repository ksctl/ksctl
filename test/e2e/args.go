// Copyright 2024 Ksctl Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ksctl/ksctl/v2/pkg/handler/cluster/controller"
	"github.com/ksctl/ksctl/v2/pkg/logger"
	"os"
)

func GetReqPayload(l logger.Logger) (Operation, controller.Metadata) {
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

	var payload controller.Metadata
	err = json.Unmarshal(raw, &payload)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	return Operation(*arg1), payload
}
