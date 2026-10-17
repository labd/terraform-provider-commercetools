package recurrence_policy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestScheduleRoundTrip(t *testing.T) {
	standard := fromSchedule(platform.StandardSchedule{Value: 2, IntervalUnit: platform.IntervalUnitWeeks})
	assert.Equal(t, []Schedule{{
		Type:         types.StringValue(Standard),
		Value:        types.Int64Value(2),
		IntervalUnit: types.StringValue("Weeks"),
		Day:          types.Int64Null(),
	}}, standard)
	assert.Equal(t,
		platform.StandardScheduleDraft{Value: 2, IntervalUnit: platform.IntervalUnitWeeks},
		standard[0].toDraft())

	monthly := fromSchedule(platform.DayOfMonthSchedule{Day: 15})
	assert.Equal(t, platform.DayOfMonthScheduleDraft{Day: 15}, monthly[0].toDraft())
	assert.True(t, monthly[0].Value.IsNull())
}

func TestToUpdateActionsSchedule(t *testing.T) {
	state := RecurrencePolicy{
		Version:  types.Int64Value(1),
		Key:      types.StringValue("monthly"),
		Schedule: fromSchedule(platform.StandardSchedule{Value: 1, IntervalUnit: platform.IntervalUnitMonths}),
	}
	assert.Empty(t, toUpdateActions(&state, &state).Actions)

	plan := state
	plan.Schedule = fromSchedule(platform.DayOfMonthSchedule{Day: 1})
	assert.Equal(t, []platform.RecurrencePolicyUpdateAction{
		platform.RecurrencePolicySetScheduleAction{Schedule: platform.DayOfMonthScheduleDraft{Day: 1}},
	}, toUpdateActions(&state, &plan).Actions)
}
