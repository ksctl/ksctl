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

package validation

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ksctl/ksctl/v2/pkg/addons"
	"github.com/ksctl/ksctl/v2/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/v2/pkg/errors"
	"github.com/ksctl/ksctl/v2/pkg/logger"
)

func ValidateDistro(distro consts.KsctlKubernetes) bool {
	if b := utf8.ValidString(string(distro)); !b {
		return false
	}

	switch distro {
	case consts.K8sK3s, consts.K8sKubeadm, "":
		return true
	default:
		return false
	}
}

func ValidateRole(role consts.KsctlRole) bool {
	if b := utf8.ValidString(string(role)); !b {
		return false
	}

	switch role {
	case consts.RoleCp, consts.RoleLb, consts.RoleWp, consts.RoleDs:
		return true
	default:
		return false
	}
}

func ValidateStorage(storage consts.KsctlStore) bool {
	if b := utf8.ValidString(string(storage)); !b {
		return false
	}

	switch storage {
	case consts.StoreExtMongo, consts.StoreLocal, consts.StoreK8s:
		return true
	default:
		return false
	}
}

func ValidateCloud(cloud consts.KsctlCloud) bool {
	if b := utf8.ValidString(string(cloud)); !b {
		return false
	}

	switch cloud {
	case consts.CloudAzure, consts.CloudAws, consts.CloudLocal, consts.CloudAll:
		return true
	default:
		return false
	}
}

func IsValidName(ctx context.Context, log logger.Logger, name string) error {
	if len(name) > 50 {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidResourceName,
			log.NewError(ctx, "name is too long", "name", name),
		)
	}
	matched, err := regexp.MatchString(`(^[a-z])([-a-z0-9])*([a-z0-9]$)`, name)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to compile the regex", "Reason", err),
		)
	}
	if !matched {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidResourceName,
			log.NewError(ctx, "invalid name", "expectedToBePattern", `(^[a-z])([-a-z0-9])*([a-z0-9]$)`),
		)
	}

	return nil
}

func IsValidKsctlComponentVersion(ctx context.Context, log logger.Logger, ver string) error {
	if ver == "latest" ||
		ver == "stable" ||
		strings.HasPrefix(ver, "feature") ||
		strings.HasPrefix(ver, "main") ||
		strings.HasPrefix(ver, "feat") ||
		strings.HasPrefix(ver, "fix") ||
		strings.HasPrefix(ver, "enhancement") {
		return nil
	}

	patternWithoutVPrefix := `^\d+(\.\d{1,2}){0,2}$`
	patternWithVPrefix := `^v\d+(\.\d{1,2}){0,2}$`
	commitShaPattern := `^\b[0-9a-f]{40}\b$`
	matchStringWithoutVPrefix, err := regexp.MatchString(patternWithoutVPrefix, ver)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to compile the regex", "Reason", err),
		)
	}
	matchStringWithVPrefix, err := regexp.MatchString(patternWithVPrefix, ver)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to compile the regex", "Reason", err),
		)
	}
	matchCommitSha, err := regexp.MatchString(commitShaPattern, ver)
	if err != nil {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrUnknown,
			log.NewError(ctx, "failed to compile the regex", "Reason", err),
		)
	}

	if !matchStringWithoutVPrefix && !matchStringWithVPrefix && !matchCommitSha {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidKsctlComponentVersion,
			log.NewError(ctx, "invalid version", "version", ver),
		)
	}
	return nil
}

func IsValidKsctlClusterAddons(ctx context.Context, log logger.Logger, ca addons.ClusterAddons) error {

	if ca == nil {
		return nil
	}

	nonKsctlCniIsNone := false
	KsctlCniIsPresent := false

	addonLabels := ca.GetAddonLabels()
	for _, label := range addonLabels {
		switch label {
		case string(consts.K8sK3s), string(consts.K8sKubeadm), string(consts.K8sKind), string(consts.K8sAks), string(consts.K8sEks), string(consts.K8sKsctl):
			_addons := ca.GetAddons(label)

			_addonNames := make(map[string]int, len(_addons))
			counterCni := 0

			for _, addon := range _addons {
				if len(addon.Name) == 0 || !utf8.ValidString(addon.Name) {
					return ksctlErrors.WrapError(
						ksctlErrors.ErrInvalidKsctlClusterAddons,
						log.NewError(ctx, "invalid addon name", "addon", addon.Name, "label", label),
					)
				}

				if addon.IsCNI {
					if addon.Name == string(consts.CNINone) {
						if label != string(consts.K8sKsctl) {
							nonKsctlCniIsNone = true
						}
					} else {
						KsctlCniIsPresent = true
					}
					counterCni++
				} else {
					_addonNames[addon.Name]++
				}
			}

			for k, v := range _addonNames {
				if v > 1 {
					return ksctlErrors.WrapError(
						ksctlErrors.ErrInvalidKsctlClusterAddons,
						log.NewError(ctx, "duplicate addon name", "addon", k, "label", label),
					)
				}
			}

			if counterCni > 1 {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidKsctlClusterAddons,
					log.NewError(ctx, "more than one cni plugin", "label", label),
				)
			}

		default:
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidKsctlClusterAddons,
				log.NewError(ctx, "invalid cluster addon", "addon", label),
			)
		}
	}

	if !nonKsctlCniIsNone && KsctlCniIsPresent {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidKsctlClusterAddons,
			log.NewError(ctx, "CONFLICT: provider specific cni is not `none` at the same time ksctl cni is there"),
		)
	}
	return nil
}
