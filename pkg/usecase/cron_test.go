package usecase

import (
	"testing"
	"time"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
)

const (
	testJobID = "test-job"
)

func TestJob_ShouldRun_CorrectTime(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)

	assert.True(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongHour(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongMinute(t *testing.T) {
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0}
	now := time.Date(2026, 3, 10, 8, 1, 0, 0, time.UTC)

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_WrongWeekday(t *testing.T) {
	sunday := time.Sunday
	job := domain.Job{ID: testJobID, Hour: 20, Minute: 0, Weekday: &sunday}
	monday := time.Date(2026, 3, 9, 20, 0, 0, 0, time.UTC) // Monday

	assert.False(t, job.ShouldRun(monday))
}

func TestJob_ShouldRun_CorrectWeekday(t *testing.T) {
	sunday := time.Sunday
	job := domain.Job{ID: testJobID, Hour: 20, Minute: 0, Weekday: &sunday}
	sun := time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC) // Sunday

	assert.True(t, job.ShouldRun(sun))
}

func TestJob_ShouldRun_TooSoonAfterLastRun(t *testing.T) {
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0, LastRun: now.Add(-30 * time.Second)}

	assert.False(t, job.ShouldRun(now))
}

func TestJob_ShouldRun_EnoughTimeAfterLastRun(t *testing.T) {
	now := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)
	job := domain.Job{ID: testJobID, Hour: 8, Minute: 0, LastRun: now.Add(-3 * time.Minute)}

	assert.True(t, job.ShouldRun(now))
}

func TestJob_Execute_WithRunFn(t *testing.T) {
	job := domain.Job{
		ID:    testJobID,
		RunFn: func() (string, error) { return "custom result", nil },
	}

	result, err := job.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "custom result", result)
}

func TestJob_Execute_NoClaude_NoRunFn_ReturnsError(t *testing.T) {
	job := domain.Job{ID: testJobID}

	_, err := job.Execute()

	assert.Error(t, err)
}

func TestJob_Deliver_LogMode(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeLog},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}

func TestJob_Deliver_WhatsApp_NotConfigured_NoError(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: domain.DeliveryModeWhatsApp},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}

func TestJob_Deliver_DefaultMode_NoError(t *testing.T) {
	job := domain.Job{
		ID:       testJobID,
		Delivery: domain.DeliveryConfig{Mode: "unknown"},
	}

	err := job.Deliver("test result")

	assert.NoError(t, err)
}
