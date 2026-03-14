package hooks

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errHookFailed = errors.New("hook failed")
)

func TestRegistry_Register_AddsHook(t *testing.T) {
	registry := NewRegistry()
	called := false

	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		called = true
		return nil
	})

	registry.Emit(context.Background(), MessageReceived, nil)

	assert.True(t, called)
}

func TestRegistry_Emit_RunsMultipleHooksSequentially(t *testing.T) {
	registry := NewRegistry()
	var order []int

	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		order = append(order, 1)
		return nil
	})
	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		order = append(order, 2)
		return nil
	})

	registry.Emit(context.Background(), MessageReceived, nil)

	assert.Equal(t, []int{1, 2}, order)
}

func TestRegistry_Emit_PassesCorrectPayload(t *testing.T) {
	registry := NewRegistry()
	var received any

	registry.Register(BeforeResponse, func(ctx context.Context, event Event) error {
		received = event.Payload
		return nil
	})

	payload := map[string]string{"key": "value"}
	registry.Emit(context.Background(), BeforeResponse, payload)

	require.NotNil(t, received)
	assert.Equal(t, payload, received)
}

func TestRegistry_Emit_SetsEventType(t *testing.T) {
	registry := NewRegistry()
	var receivedType string

	registry.Register(AfterCompaction, func(ctx context.Context, event Event) error {
		receivedType = event.Type
		return nil
	})

	registry.Emit(context.Background(), AfterCompaction, nil)

	assert.Equal(t, AfterCompaction, receivedType)
}

func TestRegistry_Emit_SetsTimestamp(t *testing.T) {
	registry := NewRegistry()
	var hasTimestamp bool

	registry.Register(CronJobCompleted, func(ctx context.Context, event Event) error {
		hasTimestamp = !event.Timestamp.IsZero()
		return nil
	})

	registry.Emit(context.Background(), CronJobCompleted, nil)

	assert.True(t, hasTimestamp)
}

func TestRegistry_Emit_ContinuesOnError(t *testing.T) {
	registry := NewRegistry()
	secondCalled := false

	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		return errHookFailed
	})
	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		secondCalled = true
		return nil
	})

	registry.Emit(context.Background(), MessageReceived, nil)

	assert.True(t, secondCalled)
}

func TestRegistry_Emit_NoHooksRegistered_DoesNotPanic(t *testing.T) {
	registry := NewRegistry()

	assert.NotPanics(t, func() {
		registry.Emit(context.Background(), "nonexistent_event", nil)
	})
}

func TestRegistry_Emit_DifferentEventsAreIsolated(t *testing.T) {
	registry := NewRegistry()
	var msgCount, cronCount int32

	registry.Register(MessageReceived, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&msgCount, 1)
		return nil
	})
	registry.Register(CronJobCompleted, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&cronCount, 1)
		return nil
	})

	registry.Emit(context.Background(), MessageReceived, nil)

	assert.Equal(t, int32(1), atomic.LoadInt32(&msgCount))
	assert.Equal(t, int32(0), atomic.LoadInt32(&cronCount))
}
