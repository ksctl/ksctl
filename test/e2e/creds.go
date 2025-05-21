// Copyright 2025 Ksctl Authors
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
	"context"
	"encoding/json"
	"os"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
)

func CredsAws(ctx context.Context) context.Context {
	accessKey, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok {
		panic("AWS_ACCESS_KEY_ID not set")
	}
	secretKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok {
		panic("AWS_SECRET_ACCESS_KEY not set")
	}

	v, err := json.Marshal(statefile.CredentialsAws{
		AccessKeyId:     accessKey,
		SecretAccessKey: secretKey,
	})
	if err != nil {
		panic(err)
	}
	return context.WithValue(ctx, consts.KsctlAwsCredentials, v)
}

func CredsAzure(ctx context.Context) context.Context {
	subscriptionId, ok := os.LookupEnv("AZURE_SUBSCRIPTION_ID")
	if !ok {
		panic("AZURE_SUBSCRIPTION_ID not set")
	}
	tenantId, ok := os.LookupEnv("AZURE_TENANT_ID")
	if !ok {
		panic("AZURE_TENANT_ID not set")
	}
	clientId, ok := os.LookupEnv("AZURE_CLIENT_ID")
	if !ok {
		panic("AZURE_CLIENT_ID not set")
	}
	clientSecret, ok := os.LookupEnv("AZURE_CLIENT_SECRET")
	if !ok {
		panic("AZURE_CLIENT_SECRET not set")
	}

	v, err := json.Marshal(statefile.CredentialsAzure{
		SubscriptionID: subscriptionId,
		TenantID:       tenantId,
		ClientID:       clientId,
		ClientSecret:   clientSecret,
	})
	if err != nil {
		panic(err)
	}
	return context.WithValue(ctx, consts.KsctlAzureCredentials, v)
}

func CredsMongo(ctx context.Context) statefile.CredentialsMongodb {

	mongoHost, ok := os.LookupEnv("MONGODB_URI")
	if !ok {
		panic("MONGODB_URI not set")
	}

	return statefile.CredentialsMongodb{
		URI: mongoHost,
	}

}
