package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsageTracker_Track(t *testing.T) {
	tracker := NewUsageTracker()

	tracker.Track("session-1", 100, 50)
	tracker.Track("session-1", 200, 100)
	tracker.Track("session-2", 50, 25)

	s1 := tracker.GetSession("session-1")
	assert.Equal(t, 300, s1.InputTokens)
	assert.Equal(t, 150, s1.OutputTokens)
	assert.Equal(t, 2, s1.Requests)

	s2 := tracker.GetSession("session-2")
	assert.Equal(t, 50, s2.InputTokens)
	assert.Equal(t, 1, s2.Requests)

	all := tracker.GetAll()
	assert.Len(t, all, 2)

	cost := tracker.TotalCostUSD()
	assert.Greater(t, cost, 0.0)
}

func TestUsageTracker_GetSession_NotFound(t *testing.T) {
	tracker := NewUsageTracker()

	s := tracker.GetSession("nonexistent")

	assert.Nil(t, s)
}

func TestSessionUsage_EstimatedCost(t *testing.T) {
	s := &SessionUsage{
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	}

	cost := s.EstimatedCostUSD()

	assert.InDelta(t, 18.0, cost, 0.01) // $3 input + $15 output
}

func TestSessionUsage_Summary(t *testing.T) {
	s := &SessionUsage{
		SessionID:    "test",
		InputTokens:  100,
		OutputTokens: 50,
		Requests:     3,
	}

	summary := s.Summary()

	assert.Contains(t, summary, "test")
	assert.Contains(t, summary, "3")
}
