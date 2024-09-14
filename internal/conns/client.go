package conns

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type Client struct {
	Config aws.Config
}

func (c *Client) Cloudfront() *cloudfront.Client {
	return cloudfront.NewFromConfig(c.Config)
}
