package service

import (
	"log"
	"sync"
)

const defaultCacheSize = 500

type cacheEntry struct {
	embedding []float64
}

// CachedEmbedder wraps an Embedder with an in-memory LRU cache.
// Identical texts return cached embeddings without an API call.
type CachedEmbedder struct {
	inner   Embedder
	mu      sync.RWMutex
	cache   map[string]cacheEntry
	order   []string
	maxSize int
	hits    int
	misses  int
}

func NewCachedEmbedder(inner Embedder, maxSize int) *CachedEmbedder {
	if maxSize <= 0 {
		maxSize = defaultCacheSize
	}
	return &CachedEmbedder{
		inner:   inner,
		cache:   make(map[string]cacheEntry, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *CachedEmbedder) Embed(text string) ([]float64, error) {
	c.mu.RLock()
	if entry, ok := c.cache[text]; ok {
		c.mu.RUnlock()
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		return entry.embedding, nil
	}
	c.mu.RUnlock()

	embedding, err := c.inner.Embed(text)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.misses++

	if len(c.cache) >= c.maxSize {
		evicted := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, evicted)
	}

	c.cache[text] = cacheEntry{embedding: embedding}
	c.order = append(c.order, text)
	c.mu.Unlock()

	if (c.hits+c.misses)%100 == 0 {
		log.Printf("embedding cache: hits=%d misses=%d size=%d", c.hits, c.misses, len(c.cache))
	}

	return embedding, nil
}

var _ Embedder = (*CachedEmbedder)(nil)
