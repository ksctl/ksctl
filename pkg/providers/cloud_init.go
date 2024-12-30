// Copyright 2024 ksctl
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

package providers

import (
	"fmt"

	"github.com/ksctl/ksctl/pkg/utilities"
)

func CloudInitScript(resName string) (string, error) {

	postfixStr, err := utilities.GenRandomString(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`#!/bin/bash
sudo hostname %s-%s

sudo cp /etc/localtime /etc/localtime.backup

sudo ln -sf /usr/share/zoneinfo/UTC /etc/localtime

`, resName, postfixStr), nil
}
