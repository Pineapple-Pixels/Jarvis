package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJob_ShouldRun_CorrectTime(t *testing.T) {
	job := Job{Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC)

	assert.True(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongHour(t *testing.T) {
	job := Job{Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongMinute(t *testing.T) {
	job := Job{Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 19, 8, 1, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongWeekday(t *testing.T) {
	sunday := time.Sunday
	job := Job{Hour: 8, Minute: 0, Weekday: &sunday}
	now := time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC) // Thursday

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_CorrectWeekday(t *testing.T) {
	thursday := time.Thursday
	job := Job{Hour: 8, Minute: 0, Weekday: &thursday}
	now := time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC) // Thursday

	assert.True(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_TooSoonAfterLastRun(t *testing.T) {
	job := Job{Hour: 8, Minute: 0, LastRun: time.Now()}
	now := time.Now()

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_EnoughTimeAfterLastRun(t *testing.T) {
	now := time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC)
	job := Job{Hour: 8, Minute: 0, LastRun: now.Add(-5 * time.Minute)}

	assert.True(t, job.ShouldRun(now))
}

func TestJob_Execute_WithRunFn(t *testing.T) {
	job := Job{
		RunFn: func() (string, error) { return "custom result", nil },
	}

	result, err := job.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "custom result", result)
}

func TestJob_Execute_NoRunFn_NoAI(t *testing.T) {
	job := Job{ID: "test-job"}

	_, err := job.Execute()

	assert.Error(t, err)
}

func TestJob_Deliver_WhatsApp_NotConfigured(t *testing.T) {
	job := Job{
		ID:       "test",
		Delivery: DeliveryConfig{Mode: DeliveryModeWhatsApp},
	}

	err := job.Deliver("result")

	assert.NoError(t, err)
}

func TestJob_Deliver_Log(t *testing.T) {
	job := Job{
		ID:       "test",
		Delivery: DeliveryConfig{Mode: DeliveryModeLog},
	}

	err := job.Deliver("result")

	assert.NoError(t, err)
}

func TestJob_Deliver_Default(t *testing.T) {
	job := Job{
		ID:       "test",
		Delivery: DeliveryConfig{Mode: "unknown"},
	}

	err := job.Deliver("result")

	assert.NoError(t, err)
}
