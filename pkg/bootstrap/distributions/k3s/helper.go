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

package k3s

import (
	"fmt"
	"github.com/ksctl/ksctl/poller"
	"strings"

	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
)

func convertK3sVersion(ver string) string {
	return fmt.Sprintf("v%s+k3s1", ver)
}

func isValidK3sVersion(ver string) (string, error) {

	validVersion, err := poller.GetSharedPoller().Get("k3s-io", "k3s")
	if err != nil {
		return "", err
	}

	if ver == "" {
		return validVersion[0], nil
	}

	ver = convertK3sVersion(ver)
	for _, vver := range validVersion {
		if vver == ver {
			return vver, nil
		}
	}
	return "", ksctlErrors.ErrInvalidVersion.Wrap(
		log.NewError(k3sCtx, "invalid k3s version", "valid versions", strings.Join(validVersion, " ")),
	)
}

func getEtcdMemberIPFieldForControlplane(ips []string) string {
	tempDS := []string{}
	for _, ip := range ips {
		newValue := fmt.Sprintf("https://%s:2379", ip)
		tempDS = append(tempDS, newValue)
	}

	return strings.Join(tempDS, ",")
}
