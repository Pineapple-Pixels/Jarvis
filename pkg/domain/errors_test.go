package domain

import (
	"fmt"
	"testing"

	stderrors "errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap_ContainsSentinelMessage(t *testing.T) {
	err := Wrap(ErrStoreOpen, "connection refused")

	assert.Equal(t, "failed to open store: connection refused", err.Error())
}

func TestWrap_IsMatchesSentinel(t *testing.T) {
	err := Wrap(ErrStoreOpen, "timeout")

	assert.True(t, stderrors.Is(err, ErrStoreOpen))
}

func TestWrap_IsDoesNotMatchDifferentSentinel(t *testing.T) {
	err := Wrap(ErrStoreOpen, "timeout")

	assert.False(t, stderrors.Is(err, ErrStoreSave))
}

func TestWrapf_WrapsWithCause(t *testing.T) {
	cause := fmt.Errorf("disk full")
	err := Wrapf(ErrStoreSave, cause)

	assert.True(t, stderrors.Is(err, ErrStoreSave))
	assert.Equal(t, "failed to save memory: disk full", err.Error())
}

func TestWrapf_UnwrapReturnsCause(t *testing.T) {
	cause := fmt.Errorf("network error")
	err := Wrapf(ErrClaudeSend, cause)

	unwrapped := stderrors.Unwrap(err)

	require.NotNil(t, unwrapped)
	assert.Equal(t, cause, unwrapped)
}

func TestWrap_UnwrapReturnsSentinel(t *testing.T) {
	err := Wrap(ErrFinanceParseExpense, "invalid json")

	unwrapped := stderrors.Unwrap(err)

	assert.Equal(t, ErrFinanceParseExpense, unwrapped)
}

func TestWrappedError_ErrorWithoutDetail(t *testing.T) {
	err := &wrappedError{sentinel: ErrClaudeEmpty}

	assert.Equal(t, "claude returned empty response", err.Error())
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		ErrStoreOpen, ErrStoreSave, ErrStoreSearch, ErrStoreFTS,
		ErrConversationLoad, ErrCompactFailed,
		ErrFinanceParseExpense, ErrFinanceWriteSheets,
		ErrClaudeAPI, ErrClaudeEmpty,
		ErrEmbedGenerate, ErrEmbedParse,
		ErrSkillsReadDir, ErrSkillsFrontmatter,
		ErrWhatsAppSend, ErrNotionRequest,
		ErrCronNoClaude,
	}

	for i := 0; i < len(sentinels); i++ {
		for j := i + 1; j < len(sentinels); j++ {
			assert.False(t, stderrors.Is(sentinels[i], sentinels[j]),
				"%v should not match %v", sentinels[i], sentinels[j])
		}
	}
}

func TestWrap_WorksWithErrorsIs_ThroughMultipleLayers(t *testing.T) {
	inner := Wrap(ErrStoreSearch, "timeout")
	middle := fmt.Errorf("memory layer: %w", inner)
	outer := fmt.Errorf("handler: %w", middle)

	assert.True(t, stderrors.Is(outer, ErrStoreSearch))
}

func TestIs_ReexportWorks(t *testing.T) {
	err := Wrap(ErrClaudeAPI, "rate limited")

	assert.True(t, Is(err, ErrClaudeAPI))
}

func TestWrapf_IsMatchesBothSentinelAndCause(t *testing.T) {
	cause := Wrap(ErrStoreOpen, "connection refused")
	err := Wrapf(ErrConversationLoad, cause)

	assert.True(t, stderrors.Is(err, ErrConversationLoad))
	assert.True(t, stderrors.Is(cause, ErrStoreOpen))
}
