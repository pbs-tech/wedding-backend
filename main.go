package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// _, err := aws.GetCallerIdentity(ctx)
		// if err != nil {
		// 	return err
		// }
		// _, err := aws.GetRegion(ctx, &aws.GetRegionArgs{})
		// if err != nil {
		// 	return err
		// }
		_, err := createDynamodbTable(ctx)
		if err != nil {
			return err
		}
		_, _, err = createLambdas(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}
