package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCloudfrontDistributionForS3(ctx *pulumi.Context, cloudfrontDistName string, distributionBucket *s3.Bucket) (*cloudfront.Distribution, error) {
	// Create Origin Access Identity (OAI)
	oai, err := cloudfront.NewOriginAccessIdentity(ctx, cloudfrontDistName+"-oai", &cloudfront.OriginAccessIdentityArgs{})
	if err != nil {
		return nil, err
	}

	// Create CloudFront Distribution
	distribution, err := cloudfront.NewDistribution(ctx, cloudfrontDistName, &cloudfront.DistributionArgs{
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				S3OriginConfig: &cloudfront.DistributionOriginS3OriginConfigArgs{
					OriginAccessIdentity: oai.CloudfrontAccessIdentityPath,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Create the S3 Bucket Policy Document
	distributionBucketPolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{
					"s3:GetObject",
				},
				Resources: []string{
					fmt.Sprintf("%v/*", distributionBucket.Arn),
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							fmt.Sprintf("%v", oai.IamArn),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	// Apply the Bucket Policy
	_, err = s3.NewBucketPolicy(ctx, "cloudfront-bucket-policy", &s3.BucketPolicyArgs{
		Bucket: distributionBucket.ID(),
		Policy: pulumi.String(distributionBucketPolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	return distribution, nil
}
