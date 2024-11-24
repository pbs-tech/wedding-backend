package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAmplifyApp(ctx *pulumi.Context) (*amplify.App, error) {
	frontEnd, err := amplify.NewApp(ctx, "wedding-frontend", &amplify.AppArgs{
		BuildSpec: pulumi.String(`version: 1
frontend:
phases:
preBuild:
commands:
- npm ci --cache .npm --prefer-offline
build:
commands:
- npm run build
artifacts:
baseDirectory: dist
files:
- '**/*'
cache:
paths:
- .npm/**/*
`),
		CacheConfig: &amplify.AppCacheConfigArgs{
			Type: pulumi.String("AMPLIFY_MANAGED"),
		},
		CustomRules: amplify.AppCustomRuleArray{
			&amplify.AppCustomRuleArgs{
				Source: pulumi.String("/<*>"),
				Status: pulumi.String("404-200"),
				Target: pulumi.String("/index.html"),
			},
			&amplify.AppCustomRuleArgs{
				Source: pulumi.String("</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|ttf|map|json)$)([^.]+$)/>"),
				Status: pulumi.String("200"),
				Target: pulumi.String("/index.html"),
			},
		},
		Name:       pulumi.String("wedding-frontend"),
		Platform:   pulumi.String("WEB"),
		Repository: pulumi.String("https://github.com/af-tec/wedding-frontend"),
	}, pulumi.Protect(true))
	if err != nil {
		return nil, err
	}
	return frontEnd, nil
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
