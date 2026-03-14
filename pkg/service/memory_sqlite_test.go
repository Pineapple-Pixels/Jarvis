//go:build cgo

package service

import (
	"os"
	"path/filepath"
	"testing"

	"asistente/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSessionID = "test-session"
	testContent   = "El pool de cartas tiene 40 cartas"
)

var (
	testTags      = []string{"game", "cards"}
	testEmbedding = []float64{0.1, 0.2, 0.3, 0.4}
)

func newTestStore(t *testing.T) *SQLiteMemoryService {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLiteMemoryService(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })

	return store
}

func TestNewStore_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLiteMemoryService(dbPath)

	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()

	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr)
}

func TestStore_Save_ReturnsID(t *testing.T) {
	store := newTestStore(t)

	id, err := store.Save(testContent, testTags, testEmbedding)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestStore_Save_IncrementingIDs(t *testing.T) {
	store := newTestStore(t)

	id1, _ := store.Save("first", nil, nil)
	id2, _ := store.Save("second", nil, nil)

	assert.Equal(t, int64(1), id1)
	assert.Equal(t, int64(2), id2)
}

func TestStore_Delete_RemovesMemory(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.Save(testContent, testTags, testEmbedding)

	err := store.Delete(id)

	require.NoError(t, err)

	results, err := store.Search(testEmbedding, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_Search_ReturnsMatchingMemories(t *testing.T) {
	store := newTestStore(t)
	store.Save("nota sobre cartas", testTags, []float64{0.9, 0.1, 0.0, 0.0})
	store.Save("nota sobre finanzas", []string{"finance"}, []float64{0.0, 0.0, 0.9, 0.1})

	results, err := store.Search([]float64{0.9, 0.1, 0.0, 0.0}, 5)

	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Contains(t, results[0].Content, "cartas")
}

func TestStore_Search_RespectsLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 10; i++ {
		store.Save("note", nil, testEmbedding)
	}

	results, err := store.Search(testEmbedding, 3)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestStore_SearchFTS_FindsByKeyword(t *testing.T) {
	store := newTestStore(t)
	store.Save("el supermercado esta caro", []string{"finance"}, nil)
	store.Save("programar en golang es divertido", []string{"code"}, nil)

	results, err := store.SearchFTS("supermercado", 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Content, "supermercado")
}

func TestStore_SearchFTS_NoMatch_ReturnsEmpty(t *testing.T) {
	store := newTestStore(t)
	store.Save("hello world", nil, nil)

	results, err := store.SearchFTS("nonexistent", 5)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_SearchHybrid_CombinesResults(t *testing.T) {
	store := newTestStore(t)
	store.Save("el auto necesita nafta", []string{"transport"}, []float64{0.8, 0.1, 0.0, 0.0})
	store.Save("comprar leche en el super", []string{"food"}, []float64{0.0, 0.0, 0.8, 0.1})

	results, err := store.SearchHybrid("nafta auto", []float64{0.8, 0.1, 0.0, 0.0}, 5, 0.6, 0.4)

	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Contains(t, results[0].Content, "nafta")
}

func TestStore_SaveConversation_And_Load(t *testing.T) {
	store := newTestStore(t)

	store.SaveConversation(testSessionID, "user", "hola")
	store.SaveConversation(testSessionID, "assistant", "hola, como estas?")

	msgs, err := store.LoadConversation(testSessionID, 10)

	require.NoError(t, err)
	require.Len(t, msgs, 2)
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "hola", msgs[0].Content)
	assert.Equal(t, "assistant", msgs[1].Role)
}

func TestStore_LoadConversation_RespectsLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 20; i++ {
		store.SaveConversation(testSessionID, "user", "msg")
	}

	msgs, err := store.LoadConversation(testSessionID, 5)

	require.NoError(t, err)
	assert.Len(t, msgs, 5)
}

func TestStore_LoadConversation_IsolatesSessions(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation("session-a", "user", "hello from a")
	store.SaveConversation("session-b", "user", "hello from b")

	msgsA, _ := store.LoadConversation("session-a", 10)
	msgsB, _ := store.LoadConversation("session-b", 10)

	assert.Len(t, msgsA, 1)
	assert.Len(t, msgsB, 1)
	assert.Equal(t, "hello from a", msgsA[0].Content)
	assert.Equal(t, "hello from b", msgsB[0].Content)
}

func TestStore_ClearConversation(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "msg")

	err := store.ClearConversation(testSessionID)

	require.NoError(t, err)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	assert.Empty(t, msgs)
}

func TestStore_ReplaceConversation(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "old message 1")
	store.SaveConversation(testSessionID, "user", "old message 2")

	replacement := []domain.ConversationMessage{
		{Role: "assistant", Content: "resumen de la conversacion"},
	}
	err := store.ReplaceConversation(testSessionID, replacement)

	require.NoError(t, err)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	require.Len(t, msgs, 1)
	assert.Equal(t, "resumen de la conversacion", msgs[0].Content)
}

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(v, v)

	assert.InDelta(t, 1.0, result, 0.0001)
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{0.0, 1.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.InDelta(t, 0.0, result, 0.0001)
}

func TestCosineSimilarity_DifferentLengths_ReturnsZero(t *testing.T) {
	a := []float64{1.0, 0.0}
	b := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_EmptyVectors_ReturnsZero(t *testing.T) {
	result := cosineSimilarity([]float64{}, []float64{})

	assert.Equal(t, 0.0, result)
}

func TestCosineSimilarity_ZeroVector_ReturnsZero(t *testing.T) {
	a := []float64{0.0, 0.0, 0.0}
	b := []float64{1.0, 0.0, 0.0}

	result := cosineSimilarity(a, b)

	assert.Equal(t, 0.0, result)
}
