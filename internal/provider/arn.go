package provider

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

var (
	_ validator.String = ArnValidator{}
)

var (
	accountIDRegexp = regexp.MustCompile(`^(aws|aws-managed|third-party|\d{12}|cw.{10})$`)
	partitionRegexp = regexp.MustCompile(`^aws(-[a-z]+)*$`)
	regionRegexp    = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d$`)
)

// ArnValidator validates that a string value matches an ARN format with additional validation on the parsed ARN value
// It must:
// * Be parseable as an ARN
// * Have a valid partition
// * Have a valid region
// * Have either an empty or valid account ID
// * Have a non-empty resource part
// * Pass the supplied checks
type ArnValidator struct {
}

func (v ArnValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("string must be a valid ARN")
}

func (v ArnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ArnValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}
	value := request.ConfigValue.ValueString()
	if value == "" {
		return
	}
	errs := v.validate(request.Path.String(), value)
	for _, err := range errs {
		response.Diagnostics.Append(diag.NewErrorDiagnostic(err.Error(), ""))
	}
}

func (v ArnValidator) validate(path, value string) []error {
	parsedARN, err := arn.Parse(value)
	if err != nil {
		return []error{fmt.Errorf("%q (%s) is an invalid ARN: %s", path, value, err.Error())}
	}

	errs := make([]error, 0)
	if parsedARN.Partition == "" {
		errs = append(errs, fmt.Errorf("%q (%s) is an invalid ARN: missing partition value", path, value))
	} else if !partitionRegexp.MatchString(parsedARN.Partition) {
		errs = append(errs, fmt.Errorf("%q (%s) is an invalid ARN: invalid partition value (expecting to match regular expression: %s)", path, value, partitionRegexp))
	}

	if parsedARN.Region != "" && !regionRegexp.MatchString(parsedARN.Region) {
		errs = append(errs, fmt.Errorf("%q (%s) is an invalid ARN: invalid region value (expecting to match regular expression: %s)", path, value, regionRegexp))
	}

	if parsedARN.AccountID != "" && !accountIDRegexp.MatchString(parsedARN.AccountID) {
		errs = append(errs, fmt.Errorf("%q (%s) is an invalid ARN: invalid account ID value (expecting to match regular expression: %s)", path, value, accountIDRegexp))
	}

	if parsedARN.Resource == "" {
		errs = append(errs, fmt.Errorf("%q (%s) is an invalid ARN: missing resource value", path, value))
	}

	return errs
}
