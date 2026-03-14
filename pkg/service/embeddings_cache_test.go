package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errEmbedInner = errors.New("inner embed failed")

type countingEmbedder struct {
	calls int
	vec   []float64
	err   error
}

func (e *countingEmbedder) Embed(text string) ([]float64, error) {
	e.calls++
	return e.vec, e.err
}

func TestCachedEmbedder_CacheHit(t *testing.T) {
	inner := &countingEmbedder{vec: []float64{0.1, 0.2}}
	cached := NewCachedEmbedder(inner, 10)

	v1, err1 := cached.Embed("hello")
	v2, err2 := cached.Embed("hello")

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, v1, v2)
	assert.Equal(t, 1, inner.calls)
}

func TestCachedEmbedder_CacheMiss(t *testing.T) {
	inner := &countingEmbedder{vec: []float64{0.1}}
	cached := NewCachedEmbedder(inner, 10)

	cached.Embed("a")
	cached.Embed("b")

	assert.Equal(t, 2, inner.calls)
}

func TestCachedEmbedder_Eviction(t *testing.T) {
	inner := &countingEmbedder{vec: []float64{0.1}}
	cached := NewCachedEmbedder(inner, 2)

	cached.Embed("a")
	cached.Embed("b")
	cached.Embed("c")

	assert.Equal(t, 3, inner.calls)

	cached.Embed("a")

	assert.Equal(t, 4, inner.calls)
}

func TestCachedEmbedder_InnerError_NotCached(t *testing.T) {
	inner := &countingEmbedder{err: errEmbedInner}
	cached := NewCachedEmbedder(inner, 10)

	_, err := cached.Embed("fail")

	require.Error(t, err)
	assert.True(t, errors.Is(err, errEmbedInner))

	inner.err = nil
	inner.vec = []float64{0.5}
	v, err := cached.Embed("fail")

	require.NoError(t, err)
	assert.Equal(t, []float64{0.5}, v)
	assert.Equal(t, 2, inner.calls)
}

func TestCachedEmbedder_DefaultSize(t *testing.T) {
	inner := &countingEmbedder{vec: []float64{0.1}}
	cached := NewCachedEmbedder(inner, 0)

	assert.Equal(t, defaultCacheSize, cached.maxSize)
}

func TestCachedEmbedder_ImplementsEmbedder(t *testing.T) {
	var _ Embedder = (*CachedEmbedder)(nil)
}
