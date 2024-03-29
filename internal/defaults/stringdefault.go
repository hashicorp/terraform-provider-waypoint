// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package defaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringDefaultValue(v types.String) planmodifier.String {
	return &stringDefaultValuePlanModifier{v}
}

type stringDefaultValuePlanModifier struct {
	DefaultValue types.String
}

var _ planmodifier.String = (*stringDefaultValuePlanModifier)(nil)

func (spm *stringDefaultValuePlanModifier) Description(ctx context.Context) string {
	/* ... */
	return ""
}

func (spm *stringDefaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	/* ... */
	return ""
}

func (spm *stringDefaultValuePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, res *planmodifier.StringResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	res.PlanValue = spm.DefaultValue
}
