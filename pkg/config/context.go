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

package config

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

func IsContextPresent(ctx context.Context, key consts.KsctlContextKeyType) (val string, isPresent bool) {
	var contextVars = [...]string{
		consts.KsctlTestFlagKey:        `true`,
		consts.KsctlModuleNameKey:      `^[\w-]+$`,
		consts.KsctlCustomDirLoc:       `^[\w-:~\\/\s]+$`,
		consts.KsctlComponentOverrides: `^([\w]+=[\w-:\.~\\/\s]+)+(,[\w]+=[\w-:\.~\\/\s]+)*$`,
	}
	_val := ctx.Value(key)
	if _val == nil {
		return "", false
	}

	switch key {
	case consts.KsctlAwsCredentials:
		if v, ok := _val.([]byte); ok {
			if err := json.Unmarshal(v, &statefile.CredentialsAws{}); err != nil {
				return "", false
			}
			return string(v), true
		} else {
			return "", false
		}
	case consts.KsctlAzureCredentials:
		if v, ok := _val.([]byte); ok {
			if err := json.Unmarshal(v, &statefile.CredentialsAzure{}); err != nil {
				return "", false
			}
			return string(v), true
		} else {
			return "", false
		}
	case consts.KsctlContextGroup:
		if v, ok := _val.(string); ok {
			return v, true
		} else {
			return "", false
		}
	case consts.KsctlContextUser:
		if v, ok := _val.(string); ok {
			return v, true
		} else {
			return "", false
		}
	default:
		expectedPattern := contextVars[key]

		gotV, ok := _val.(string)
		if ok {
			_ok, err := regexp.MatchString(expectedPattern, gotV)
			if err != nil {
				return "", false
			}
			if _ok {
				return gotV, true
			}
		}
	}

	return "", false
}
