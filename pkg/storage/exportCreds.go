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

package storage

import (
	"context"

	"github.com/ksctl/ksctl/internal/storage/external/mongodb"
	"github.com/ksctl/ksctl/pkg/helpers/consts"
	ksctlErrors "github.com/ksctl/ksctl/pkg/helpers/errors"
	"github.com/ksctl/ksctl/pkg/types"
)

func HandleCreds(ctx context.Context, log types.LoggerFactory, store consts.KsctlStore) (map[string][]byte, error) {
	switch store {
	case consts.StoreLocal, consts.StoreK8s:
		return nil, ksctlErrors.ErrInvalidStorageProvider.Wrap(
			log.NewError(ctx, "these are not external storageProvider"),
		)
	case consts.StoreExtMongo:
		return mongodb.ExportEndpoint()
	default:
		return nil, ksctlErrors.ErrInvalidStorageProvider.Wrap(
			log.NewError(ctx, "invalid storage", "storage", store),
		)
	}
}
