package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// testAccMockCloudfront serves a minimal in-process CloudFront API and points the AWS SDK
// at it via environment variables, so acceptance tests can exercise the full resource
// lifecycle without AWS credentials or real infrastructure. Both the provider (via
// aws-sdk-go-base) and testAccCreateCdn build their config from the environment, so the
// AWS_ENDPOINT_URL_CLOUDFRONT override reaches every client.
//
// Invalidations report "InProgress" on create and "Completed" on the first poll, which
// satisfies the InvalidationCompletedWaiter on its immediate first attempt.
func testAccMockCloudfront(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		w.Header().Set("Content-Type", "text/xml")
		switch {
		case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "distribution":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `<Distribution><Id>%s</Id><Status>Deployed</Status><DomainName>%s.cloudfront.example</DomainName></Distribution>`,
				mockCfId("E"), strings.ToLower(mockCfId("d")))
		case r.Method == http.MethodDelete && len(parts) == 3 && parts[1] == "distribution":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && len(parts) == 4 && parts[3] == "invalidation":
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `<Invalidation><Id>%s</Id><Status>InProgress</Status></Invalidation>`, mockCfId("I"))
		case r.Method == http.MethodGet && len(parts) == 5 && parts[3] == "invalidation":
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `<Invalidation><Id>%s</Id><Status>Completed</Status></Invalidation>`, parts[4])
		default:
			t.Errorf("mock cloudfront: unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	// aws-sdk-go-base validates credentials with sts:GetCallerIdentity during provider
	// Configure, so STS needs a mock as well.
	stsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::123456789012:user/mock</Arn>
    <UserId>AIDAMOCKMOCKMOCKMOCK</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata><RequestId>00000000-0000-0000-0000-000000000000</RequestId></ResponseMetadata>
</GetCallerIdentityResponse>`)
	}))
	t.Cleanup(stsSrv.Close)

	t.Setenv("AWS_ENDPOINT_URL_CLOUDFRONT", srv.URL)
	t.Setenv("AWS_ENDPOINT_URL_STS", stsSrv.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "mock")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "mock")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}

func mockCfId(prefix string) string {
	return (prefix + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")))[:14]
}
