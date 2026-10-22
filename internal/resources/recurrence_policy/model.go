package recurrence_policy

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

const (
	Standard   = "standard"
	DayOfMonth = "dayOfMonth"
)

// RecurrencePolicy is the main resource schema data
type RecurrencePolicy struct {
	ID          types.String                     `tfsdk:"id"`
	Version     types.Int64                      `tfsdk:"version"`
	Key         types.String                     `tfsdk:"key"`
	Name        customtypes.LocalizedStringValue `tfsdk:"name"`
	Description customtypes.LocalizedStringValue `tfsdk:"description"`
	// One schedule, as a list for the reason the project resource gives.
	Schedule []Schedule `tfsdk:"schedule"`
}

type Schedule struct {
	Type         types.String `tfsdk:"type"`
	Value        types.Int64  `tfsdk:"value"`
	IntervalUnit types.String `tfsdk:"interval_unit"`
	Day          types.Int64  `tfsdk:"day"`
}

func (s Schedule) toDraft() platform.RecurrencePolicyScheduleDraft {
	if s.Type.ValueString() == DayOfMonth {
		return platform.DayOfMonthScheduleDraft{Day: int(s.Day.ValueInt64())}
	}
	return platform.StandardScheduleDraft{
		Value:        int(s.Value.ValueInt64()),
		IntervalUnit: platform.IntervalUnit(s.IntervalUnit.ValueString()),
	}
}

func toDraft(p *RecurrencePolicy) platform.RecurrencePolicyDraft {
	return platform.RecurrencePolicyDraft{
		Key:         p.Key.ValueString(),
		Name:        p.Name.ValueLocalizedStringRef(),
		Description: p.Description.ValueLocalizedStringRef(),
		Schedule:    p.Schedule[0].toDraft(),
	}
}

func fromSchedule(schedule platform.RecurrencePolicySchedule) []Schedule {
	empty := Schedule{
		Value:        types.Int64Null(),
		IntervalUnit: types.StringNull(),
		Day:          types.Int64Null(),
	}
	switch s := schedule.(type) {
	case platform.StandardSchedule:
		empty.Type = types.StringValue(Standard)
		empty.Value = types.Int64Value(int64(s.Value))
		empty.IntervalUnit = types.StringValue(string(s.IntervalUnit))
	case *platform.StandardSchedule:
		empty.Type = types.StringValue(Standard)
		empty.Value = types.Int64Value(int64(s.Value))
		empty.IntervalUnit = types.StringValue(string(s.IntervalUnit))
	case platform.DayOfMonthSchedule:
		empty.Type = types.StringValue(DayOfMonth)
		empty.Day = types.Int64Value(int64(s.Day))
	case *platform.DayOfMonthSchedule:
		empty.Type = types.StringValue(DayOfMonth)
		empty.Day = types.Int64Value(int64(s.Day))
	default:
		return nil
	}
	return []Schedule{empty}
}

func fromNative(n *platform.RecurrencePolicy) RecurrencePolicy {
	return RecurrencePolicy{
		ID:          types.StringValue(n.ID),
		Version:     types.Int64Value(int64(n.Version)),
		Key:         types.StringValue(n.Key),
		Name:        utils.FromOptionalLocalizedString(n.Name),
		Description: utils.FromOptionalLocalizedString(n.Description),
		Schedule:    fromSchedule(n.Schedule),
	}
}

func toUpdateActions(state *RecurrencePolicy, plan *RecurrencePolicy) platform.RecurrencePolicyUpdate {
	update := platform.RecurrencePolicyUpdate{
		Version: int(state.Version.ValueInt64()),
		Actions: []platform.RecurrencePolicyUpdateAction{},
	}

	if !state.Key.Equal(plan.Key) {
		update.Actions = append(update.Actions, platform.RecurrencePolicySetKeyAction{
			Key: plan.Key.ValueStringPointer(),
		})
	}

	if !reflect.DeepEqual(state.Name, plan.Name) {
		update.Actions = append(update.Actions, platform.RecurrencePolicySetNameAction{
			Name: plan.Name.ValueLocalizedStringRef(),
		})
	}

	if !reflect.DeepEqual(state.Description, plan.Description) {
		update.Actions = append(update.Actions, platform.RecurrencePolicySetDescriptionAction{
			Description: plan.Description.ValueLocalizedStringRef(),
		})
	}

	if !reflect.DeepEqual(state.Schedule, plan.Schedule) {
		update.Actions = append(update.Actions, platform.RecurrencePolicySetScheduleAction{
			Schedule: plan.Schedule[0].toDraft(),
		})
	}

	return update
}
