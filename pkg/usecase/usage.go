package usecase

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// UsageTracker tracks token consumption and estimated costs per session.
type UsageTracker struct {
	mu       sync.Mutex
	sessions map[string]*SessionUsage
}

// SessionUsage holds cumulative token usage for a session.
type SessionUsage struct {
	SessionID    string    `json:"session_id"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Requests     int       `json:"requests"`
	FirstUsed    time.Time `json:"first_used"`
	LastUsed     time.Time `json:"last_used"`
}

// EstimatedCostUSD returns the estimated cost in USD.
// Uses Claude Sonnet 4 pricing: $3/MTok input, $15/MTok output.
func (s *SessionUsage) EstimatedCostUSD() float64 {
	inputCost := float64(s.InputTokens) / 1_000_000 * 3.0
	outputCost := float64(s.OutputTokens) / 1_000_000 * 15.0
	return inputCost + outputCost
}

// Summary returns a human-readable usage summary.
func (s *SessionUsage) Summary() string {
	return fmt.Sprintf("Sesión: %s | Requests: %d | Tokens: %d in / %d out | Costo: ~US$%.4f",
		s.SessionID, s.Requests, s.InputTokens, s.OutputTokens, s.EstimatedCostUSD())
}

// NewUsageTracker creates a new usage tracker.
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		sessions: make(map[string]*SessionUsage),
	}
}

// Track records token usage for a session.
func (u *UsageTracker) Track(sessionID string, inputTokens, outputTokens int) {
	u.mu.Lock()
	defer u.mu.Unlock()

	s, ok := u.sessions[sessionID]
	if !ok {
		s = &SessionUsage{
			SessionID: sessionID,
			FirstUsed: time.Now(),
		}
		u.sessions[sessionID] = s
	}

	s.InputTokens += inputTokens
	s.OutputTokens += outputTokens
	s.Requests++
	s.LastUsed = time.Now()

	log.Printf("usage: session=%s req=%d in=%d out=%d total_cost=US$%.4f",
		sessionID, s.Requests, s.InputTokens, s.OutputTokens, s.EstimatedCostUSD())
}

// GetSession returns usage for a specific session.
func (u *UsageTracker) GetSession(sessionID string) *SessionUsage {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.sessions[sessionID]
}

// GetAll returns all session usage data.
func (u *UsageTracker) GetAll() []SessionUsage {
	u.mu.Lock()
	defer u.mu.Unlock()

	result := make([]SessionUsage, 0, len(u.sessions))
	for _, s := range u.sessions {
		result = append(result, *s)
	}
	return result
}

// TotalCostUSD returns the total estimated cost across all sessions.
func (u *UsageTracker) TotalCostUSD() float64 {
	u.mu.Lock()
	defer u.mu.Unlock()

	total := 0.0
	for _, s := range u.sessions {
		total += s.EstimatedCostUSD()
	}
	return total
}
