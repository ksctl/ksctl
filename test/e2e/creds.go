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
	"strconv"

	"github.com/ksctl/ksctl/v2/pkg/consts"
	"github.com/ksctl/ksctl/v2/pkg/statefile"
	"github.com/ksctl/ksctl/v2/pkg/utilities"
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

func CredsMongo(ctx context.Context) context.Context {
	var mongodbSrv bool
	var mongoPort int

	if mongoSchema, ok := os.LookupEnv("MONGODB_SRV"); !ok {
		mongodbSrv = false
	} else {
		if v, err := strconv.ParseBool(mongoSchema); err != nil {
			panic("MONGODB_SRV must be a boolean")
		} else {
			mongodbSrv = v
		}
	}

	mongoHost, ok := os.LookupEnv("MONGODB_HOST")
	if !ok {
		panic("MONGODB_HOST not set")
	}

	if v, ok := os.LookupEnv("MONGODB_PORT"); !ok {
		mongoPort = 27017
	} else {
		if v, err := strconv.Atoi(v); err != nil {
			panic("MONGODB_PORT must be an integer")
		} else {
			mongoPort = v
		}
	}

	mongoUser, ok := os.LookupEnv("MONGODB_USER")
	if !ok {
		panic("MONGODB_USER not set")
	}

	mongoPass, ok := os.LookupEnv("MONGODB_PASS")
	if !ok {
		panic("MONGODB_PASS not set")
	}

	v, err := json.Marshal(statefile.CredentialsMongodb{
		SRV:      mongodbSrv,
		Username: mongoUser,
		Password: mongoPass,
		Domain:   mongoHost,
		Port:     utilities.Ptr(mongoPort),
	})
	if err != nil {
		panic(err)
	}
	return context.WithValue(ctx, consts.KsctlMongodbCredentials, v)
}
