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
	"slices"
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

func ValidateClusterType(clusterType consts.KsctlClusterType) bool {
	if b := utf8.ValidString(string(clusterType)); !b {
		return false
	}

	switch clusterType {
	case consts.ClusterTypeSelfMang, consts.ClusterTypeMang:
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

	if len(ca) == 0 {
		return nil
	}

	addonLabels := ca.GetAddonLabels()

	if len(addonLabels) > 2 {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidKsctlClusterAddons,
			log.NewError(ctx, "more than 2 labels are present in the cluster addons"),
		)
	}

	if len(addonLabels) == 2 && !slices.Contains(addonLabels, string(consts.K8sKsctl)) {
		return ksctlErrors.WrapError(
			ksctlErrors.ErrInvalidKsctlClusterAddons,
			log.NewError(ctx, "ksctl label is not present in the cluster addons"),
		)
	}

	cniCounter := 0
	foundNonKsctlAddonHavingNoneCni := false
	foundKsctlAddonHavingCni := false

	validLabels := []string{string(consts.K8sAks), string(consts.K8sEks), string(consts.K8sK3s), string(consts.K8sKind), string(consts.K8sKubeadm), string(consts.K8sKsctl)}

	for _, label := range addonLabels {
		_addons := ca.GetAddons(label)

		if !slices.Contains(validLabels, label) {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidKsctlClusterAddons,
				log.NewError(ctx, "invalid label", "label", label),
			)
		}

		_addonNames := make(map[string]int, len(_addons))
		for _, addon := range _addons {
			if addon.Name == "" || !utf8.ValidString(addon.Name) {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidKsctlClusterAddons,
					log.NewError(ctx, "invalid addon name", "addon", addon.Name, "label", label),
				)
			}

			if !addon.IsCNI {
				_addonNames[addon.Name]++
			} else {
				cniCounter++
				if addon.Name == string(consts.CNINone) {
					if label == string(consts.K8sKsctl) {
						return ksctlErrors.WrapError(
							ksctlErrors.ErrInvalidKsctlClusterAddons,
							log.NewError(ctx, "ksctl cni is `none` which is not possible"),
						)
					} else {
						foundNonKsctlAddonHavingNoneCni = true
						// mark that non ksctl cni is none
					}
				} else {
					if label == string(consts.K8sKsctl) {
						foundKsctlAddonHavingCni = true
					}
				}
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
	}

	if foundNonKsctlAddonHavingNoneCni {
		if foundKsctlAddonHavingCni {
			if cniCounter != 2 {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidKsctlClusterAddons,
					log.NewError(ctx, "only 2 cni plugins should be present"),
				)
			}
		} else {
			if cniCounter != 1 {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidKsctlClusterAddons,
					log.NewError(ctx, "only one cni plugin should be present"),
				)
			}
		}
	} else {
		if foundKsctlAddonHavingCni {
			return ksctlErrors.WrapError(
				ksctlErrors.ErrInvalidKsctlClusterAddons,
				log.NewError(ctx, "ksctl cni is present but non ksctl cni is not `none`"),
			)
		} else {
			if cniCounter != 1 {
				return ksctlErrors.WrapError(
					ksctlErrors.ErrInvalidKsctlClusterAddons,
					log.NewError(ctx, "only one cni plugin should be present"),
				)
			}
		}
	}

	return nil
}
