package cloudfront

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-awsex/internal/conns"
	"time"
)

func CreateInvalidation(ctx context.Context, client *conns.Client, distributionId string, paths []string, createTimeout time.Duration) (*cftypes.Invalidation, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionId,
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: aws.String(uuid.NewString()),
			Paths: &cftypes.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}
	cfClient := client.Cloudfront()
	out, err := cfClient.CreateInvalidation(ctx, input)
	if err != nil {
		diags.AddError("Error creating AWS Cloudfront Invalidation", err.Error())
		return nil, diags
	}
	if out != nil && out.Invalidation != nil && out.Invalidation.Id != nil {
		tflog.Trace(ctx, "Created Cloudfront Invalidation")
	} else {
		diags.AddWarning("Unable to create AWS Cloudfront Invalidation.", "AWS did not create an invalidation and gave no reason")
		return nil, diags
	}

	waiter := cloudfront.NewInvalidationCompletedWaiter(cfClient)
	res, err := waiter.WaitForOutput(ctx, &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(distributionId),
		Id:             out.Invalidation.Id,
	}, createTimeout)
	if err != nil {
		diags.AddError("Error waiting for creation of AWS Cloudfront Invalidation", err.Error())
		return out.Invalidation, diags
	} else if res != nil && res.Invalidation != nil {
		return res.Invalidation, diags
	}

	return out.Invalidation, diags
}

func FindInvalidation(ctx context.Context, client *conns.Client, distributionId string, id string) (*cftypes.Invalidation, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &cloudfront.GetInvalidationInput{
		DistributionId: &distributionId,
		Id:             &id,
	}
	out, err := client.Cloudfront().GetInvalidation(ctx, input)
	if err != nil {
		var nsi *cftypes.NoSuchInvalidation
		if !errors.As(err, &nsi) {
			diags.AddError("error getting AWS Invalidation", err.Error())
		}
	}
	if out != nil {
		return out.Invalidation, diags
	}
	return nil, diags
}
