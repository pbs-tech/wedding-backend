package main

import (
	"github.com/pulumi/pulumi-aws/sdk/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createS3Bucket(ctx *pulumi.Context, bucketName string) (*s3.Bucket, error) {
	bucket, err := s3.NewBucketV2(ctx, bucketName, &s3.BucketArgs{
		Bucket:       pulumi.String(bucketName),
		ForceDestroy: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("My bucket"),
		},
	})
	if err != nil {
		return nil, err
	}
	_, err = s3.NewBucketAclV2(ctx, bucketName+"-acl", &s3.BucketAclV2Args{
		Bucket: bucket.ID(),
		Acl:    pulumi.String("private"),
	})
	if err != nil {
		return err
	}
	return bucket, err
}
