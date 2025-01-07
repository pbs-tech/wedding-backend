package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createCloudfrontDistributionForApiGateway(ctx *pulumi.Context, cloudfrontDistName string, apiGateway *apigatewayv2.Api) (*cloudfront.Distribution, error) {
	// Create Origin Access Identity (OAI)
	oai, err := cloudfront.NewOriginAccessIdentity(ctx, cloudfrontDistName+"-oai", &cloudfront.OriginAccessIdentityArgs{})
	if err != nil {
		return nil, err
	}

	// Create CloudFront Distribution
	distribution, err := cloudfront.NewDistribution(ctx, cloudfrontDistName, &cloudfront.DistributionArgs{
		Enabled: pulumi.Bool(true),
		Origins: cloudfront.DistributionOriginArray{
			&cloudfront.DistributionOriginArgs{
				DomainName: apiGateway.ApiEndpoint,
				OriginId:   oai.ID(),
				CustomOriginConfig: &cloudfront.DistributionOriginCustomOriginConfigArgs{
					HttpPort:             pulumi.Int(80),
					HttpsPort:            pulumi.Int(443),
					OriginProtocolPolicy: pulumi.String("https-only"),
				},
			},
		},
		DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
			TargetOriginId:       oai.ID(),
			ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
			AllowedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("POST"),
				pulumi.String("OPTIONS"),
			},
			CachedMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("POST"),
				pulumi.String("OPTIONS"),
			},
			ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
				QueryString: pulumi.Bool(true),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return distribution, nil
}
