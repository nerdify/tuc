package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	cfg, _ = external.LoadDefaultAWSConfig()
	svc    = dynamodb.New(cfg)
)
