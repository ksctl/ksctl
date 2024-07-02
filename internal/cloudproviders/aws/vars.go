package aws

import (
	"context"

	"github.com/ksctl/ksctl/pkg/helpers/consts"
	"github.com/ksctl/ksctl/pkg/types"
	storageTypes "github.com/ksctl/ksctl/pkg/types/storage"
)

var (
	mainStateDocument *storageTypes.StorageDocument
	clusterType       consts.KsctlClusterType
	log               types.LoggerFactory
	awsCtx            context.Context
)

const (
	assumeClusterRolePolicyDocument = `{
    "Version": "2012-10-17",
    "Statement": {
        "Sid": "TrustPolicyStatementThatAllowsEC2ServiceToAssumeTheAttachedRole",
        "Effect": "Allow",
        "Principal": { "Service": "eks.amazonaws.com" },
       "Action": "sts:AssumeRole"
    }
}                    
  `

	assumeWorkerNodeRolePolicyDocument = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}`
)
