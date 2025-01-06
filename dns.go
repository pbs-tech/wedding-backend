package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createDnsZone(ctx *pulumi.Context, domainName string) (*route53.Zone, error) {
	zone, err := route53.NewZone(ctx, domainName, &route53.ZoneArgs{
		Name: pulumi.String(domainName),
	}, pulumi.Protect(true))
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func configureDnsForApiGateway(ctx *pulumi.Context, apiDomainStr string, zoneId pulumi.StringOutput) (*apigatewayv2.DomainName, error) {

	// Request ACM cert
	sslCert, err := acm.NewCertificate(ctx,
		"ssl-cert",
		&acm.CertificateArgs{
			DomainName:       pulumi.String(apiDomainStr),
			ValidationMethod: pulumi.String("DNS"),
		},
	)
	if err != nil {
		return nil, err
	}
	domainValidationOption := sslCert.DomainValidationOptions.ApplyT(func(options []acm.CertificateDomainValidationOption) acm.CertificateDomainValidationOption {
		return options[0]
	})

	// Create DNS record
	sslCertValidationDnsRecord, err := route53.NewRecord(ctx,
		"ssl-cert-validation-dns-record",
		&route53.RecordArgs{
			ZoneId: zoneId,
			Name: domainValidationOption.ApplyT(func(option acm.CertificateDomainValidationOption) string {
				return *option.ResourceRecordName
			}).(pulumi.StringOutput),
			Type: domainValidationOption.ApplyT(func(option acm.CertificateDomainValidationOption) string {
				return *option.ResourceRecordType
			}).(pulumi.StringOutput),
			Records: pulumi.StringArray{
				domainValidationOption.ApplyT(func(option acm.CertificateDomainValidationOption) string {
					return *option.ResourceRecordValue
				}).(pulumi.StringOutput),
			},
			Ttl: pulumi.Int(10 * 60), // 10 mins
		},
	)
	if err != nil {
		return nil, err
	}

	// wait for cert validation
	validatedSslCert, err := acm.NewCertificateValidation(ctx,
		"ssl-cert-validation",
		&acm.CertificateValidationArgs{
			CertificateArn:        sslCert.Arn,
			ValidationRecordFqdns: pulumi.StringArray{sslCertValidationDnsRecord.Fqdn},
		},
	)
	if err != nil {
		return nil, err
	}

	// configure apigw v2 with domain name and cert
	apiDomainName, err := apigatewayv2.NewDomainName(ctx, "api-domain-name",
		&apigatewayv2.DomainNameArgs{
			DomainName: pulumi.String(apiDomainStr),
			DomainNameConfiguration: &apigatewayv2.DomainNameDomainNameConfigurationArgs{
				CertificateArn: validatedSslCert.CertificateArn,
				EndpointType:   pulumi.String("REGIONAL"),
				SecurityPolicy: pulumi.String("TLS_1_2"),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return apiDomainName, nil
}

func mapDnsToApiGateway(ctx *pulumi.Context, apiDomainStr string, apiDomainName *apigatewayv2.DomainName, apiStageId pulumi.IDOutput, apiGatewayId pulumi.IDOutput, zoneId pulumi.StringOutput) error {
	// Configure domain mapping: Associate the domain with the API stage
	_, err := apigatewayv2.NewApiMapping(ctx,
		"api-domain-mapping",
		&apigatewayv2.ApiMappingArgs{
			ApiId:      apiGatewayId,
			DomainName: apiDomainName.DomainName,
			Stage:      apiStageId,
		},
	)
	if err != nil {
		return err
	}

	// Use the Hosted Zone ID of the API Gateway Domain Name (not the newly created Zone)
	apiDomainHostedZoneId := apiDomainName.DomainNameConfiguration.HostedZoneId().Elem()
	apiDomainNameTargetDomainName := apiDomainName.DomainNameConfiguration.TargetDomainName().Elem()
	_, err = route53.NewRecord(ctx, "api-route53-a-record", &route53.RecordArgs{
		Aliases: route53.RecordAliasArray{
			&route53.RecordAliasArgs{
				EvaluateTargetHealth: pulumi.Bool(false),
				Name:                 apiDomainNameTargetDomainName,
				ZoneId:               apiDomainHostedZoneId},
		},
		Name:   pulumi.String(apiDomainStr),
		Type:   pulumi.String(route53.RecordTypeA),
		ZoneId: zoneId,
	}, pulumi.Protect(true))

	if err != nil {
		return err
	}
	return nil
}

func createAmplifyDomain(ctx *pulumi.Context, frontEnd *amplify.App, domain string) (*amplify.DomainAssociation, error) {
	frontendDomain, err := amplify.NewDomainAssociation(ctx, "wedding-frontend-domain", &amplify.DomainAssociationArgs{
		AppId: frontEnd.ID(),
		CertificateSettings: &amplify.DomainAssociationCertificateSettingsArgs{
			Type: pulumi.String("AMPLIFY_MANAGED"),
		},
		DomainName: pulumi.String(domain),
		SubDomains: amplify.DomainAssociationSubDomainArray{
			&amplify.DomainAssociationSubDomainArgs{
				BranchName: pulumi.String("main"),
				Prefix:     pulumi.String(""),
			},
			&amplify.DomainAssociationSubDomainArgs{
				BranchName: pulumi.String("main"),
				Prefix:     pulumi.String("www"),
			},
		},
	}, pulumi.Protect(true))
	if err != nil {
		return frontendDomain, err
	}
	return frontendDomain, err
}
