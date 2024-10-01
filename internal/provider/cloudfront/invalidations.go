package cloudfront

import (
	"context"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-awsex/internal/conns"
	"sync"
	"time"
)

type invalidationResult struct {
	DistributionId string
	Invalidation   *cftypes.Invalidation
	Diags          diag.Diagnostics
}

func CreateInvalidations(ctx context.Context, client *conns.Client, distributionIds []string, paths []string,
	createTimeout time.Duration) (map[string]*cftypes.Invalidation, diag.Diagnostics) {
	ch := make(chan invalidationResult, len(distributionIds))

	var wg sync.WaitGroup
	for _, distributionId := range distributionIds {
		wg.Add(1)
		go func(distributionId string) {
			defer wg.Done()
			inval, diags := CreateInvalidation(ctx, client, distributionId, paths, createTimeout)
			ch <- invalidationResult{
				DistributionId: distributionId,
				Invalidation:   inval,
				Diags:          diags,
			}

		}(distributionId)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make(map[string]*cftypes.Invalidation)
	var diags diag.Diagnostics
	for cur := range ch {
		results[cur.DistributionId] = cur.Invalidation
		diags.Append(cur.Diags...)
	}
	return results, diags
}

func FindInvalidations(ctx context.Context, client *conns.Client, ids map[string]string) (map[string]*cftypes.Invalidation, diag.Diagnostics) {
	ch := make(chan invalidationResult, len(ids))

	var wg sync.WaitGroup
	for distributionId, id := range ids {
		wg.Add(1)
		go func(distributionId, id string) {
			defer wg.Done()
			inval, diags := FindInvalidation(ctx, client, distributionId, id)
			ch <- invalidationResult{
				DistributionId: distributionId,
				Invalidation:   inval,
				Diags:          diags,
			}

		}(distributionId, id)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make(map[string]*cftypes.Invalidation)
	var diags diag.Diagnostics
	for cur := range ch {
		results[cur.DistributionId] = cur.Invalidation
		diags.Append(cur.Diags...)
	}
	return results, diags
}
