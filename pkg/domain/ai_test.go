package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithMaxTokens(t *testing.T) {
	opt := WithMaxTokens(4096)
	cfg := &CompletionConfig{}

	opt(cfg)

	assert.Equal(t, 4096, cfg.MaxTokens)
}

func TestApplyOptions_Defaults(t *testing.T) {
	cfg := ApplyOptions(2048)

	assert.Equal(t, 2048, cfg.MaxTokens)
}

func TestApplyOptions_WithOverride(t *testing.T) {
	cfg := ApplyOptions(2048, WithMaxTokens(8192))

	assert.Equal(t, 8192, cfg.MaxTokens)
}

func TestApplyOptions_MultipleOptions(t *testing.T) {
	cfg := ApplyOptions(1000, WithMaxTokens(2000), WithMaxTokens(3000))

	assert.Equal(t, 3000, cfg.MaxTokens)
}
