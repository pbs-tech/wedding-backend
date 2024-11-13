package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func configureDns(ctx *pulumi.Context, domain string, zoneId string) (*apigatewayv2.DomainName, error) {
	// Request ACM cert
	sslCert, err := acm.NewCertificate(ctx,
		"ssl-cert",
		&acm.CertificateArgs{
			DomainName:       pulumi.String(domain),
			ValidationMethod: pulumi.String("DNS"),
		},
	)
	if err != nil {
		return nil, err
	}
	domainValidationOption := sslCert.DomainValidationOptions.ApplyT(func(options []acm.CertificateDomainValidationOption) interface{} {
		return options[0]
	})
	// Create DNS record
	sslCertValidationDnsRecord, err := route53.NewRecord(ctx,
		"ssl-cert-validation-dns-record",
		&route53.RecordArgs{
			ZoneId: pulumi.String(zoneId),
			Name: domainValidationOption.ApplyT(func(option interface{}) string {
				return *option.(acm.CertificateDomainValidationOption).ResourceRecordName
			}).(pulumi.StringOutput),
			Type: domainValidationOption.ApplyT(func(option interface{}) string {
				return *option.(acm.CertificateDomainValidationOption).ResourceRecordType
			}).(pulumi.StringOutput),
			Records: pulumi.StringArray{
				domainValidationOption.ApplyT(func(option interface{}) string {
					return *option.(acm.CertificateDomainValidationOption).ResourceRecordValue
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
			DomainName: pulumi.String(domain),
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

	// Create DNS record
	_, err = route53.NewRecord(ctx, "api-route53-record",
		&route53.RecordArgs{
			ZoneId: pulumi.String(zoneId),
			Type:   pulumi.String("A"),
			Name:   pulumi.String(domain),
			Aliases: route53.RecordAliasArray{
				route53.RecordAliasArgs{
					Name:                 apiDomainName.DomainName,
					EvaluateTargetHealth: pulumi.Bool(false),
				},
			},
		})
	if err != nil {
		return nil, err
	}
	return apiDomainName, nil
}
