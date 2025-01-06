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
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ksctl/ksctl/pkg/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/errors"
	"github.com/ksctl/ksctl/pkg/logger"
	"golang.org/x/term"
)

func UserInputCredentials(ctx context.Context, logging logger.Logger) (string, error) {

	logging.Print(ctx, "Enter Secret")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	if len(bytePassword) == 0 {
		logging.Error("Empty secret passed!")
		return UserInputCredentials(ctx, logging)
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

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

func ValidCNIPlugin(cni consts.KsctlValidCNIPlugin) bool {

	if b := utf8.ValidString(string(cni)); !b {
		return false
	}

	switch cni {
	case consts.CNIAzure, consts.CNICilium, consts.CNIFlannel, consts.CNIKubenet, consts.CNIKind, "":
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
