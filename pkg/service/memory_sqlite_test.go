//go:build cgo

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

// --------------- NewSQLiteMemoryService ---------------

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

func TestNewStore_InvalidPath_ReturnsError(t *testing.T) {
	store, err := NewSQLiteMemoryService("/nonexistent/path/that/does/not/exist/test.db")

	assert.Error(t, err)
	assert.Nil(t, store)
}

func TestNewStore_ImplementsMemoryService(t *testing.T) {
	var _ MemoryService = (*SQLiteMemoryService)(nil)
}

// --------------- Save ---------------

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

func TestStore_Save_NilTagsAndEmbedding(t *testing.T) {
	store := newTestStore(t)

	id, err := store.Save("content with nils", nil, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestStore_Save_EmptyContent(t *testing.T) {
	store := newTestStore(t)

	id, err := store.Save("", []string{}, []float64{})

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

func TestStore_Save_LargeEmbedding(t *testing.T) {
	store := newTestStore(t)
	emb := make([]float64, 64)
	for i := range emb {
		emb[i] = float64(i) / 64.0
	}

	id, err := store.Save("large embedding test", []string{"test"}, emb)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
}

// --------------- Delete ---------------

func TestStore_Delete_RemovesMemory(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.Save(testContent, testTags, testEmbedding)

	err := store.Delete(id)

	require.NoError(t, err)

	results, err := store.Search(testEmbedding, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_Delete_NonexistentID_NoError(t *testing.T) {
	store := newTestStore(t)

	err := store.Delete(999)

	assert.NoError(t, err)
}

func TestStore_Delete_DoesNotAffectOtherMemories(t *testing.T) {
	store := newTestStore(t)
	id1, _ := store.Save("first", nil, testEmbedding)
	store.Save("second", nil, testEmbedding)

	store.Delete(id1)

	results, err := store.Search(testEmbedding, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "second", results[0].Content)
}

// --------------- Search (vector) ---------------

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

func TestStore_Search_EmptyStore_ReturnsEmpty(t *testing.T) {
	store := newTestStore(t)

	results, err := store.Search(testEmbedding, 5)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_Search_SetsScoreField(t *testing.T) {
	store := newTestStore(t)
	store.Save("test content", nil, []float64{1.0, 0.0, 0.0, 0.0})

	results, err := store.Search([]float64{1.0, 0.0, 0.0, 0.0}, 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Greater(t, results[0].Score, 0.0)
}

func TestStore_Search_OrdersByScore(t *testing.T) {
	store := newTestStore(t)
	store.Save("low match", nil, []float64{0.0, 0.0, 0.0, 1.0})
	store.Save("high match", nil, []float64{1.0, 0.0, 0.0, 0.0})
	store.Save("medium match", nil, []float64{0.5, 0.5, 0.0, 0.0})

	results, err := store.Search([]float64{1.0, 0.0, 0.0, 0.0}, 3)

	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "high match", results[0].Content)
}

func TestStore_Search_LimitExceedsResults(t *testing.T) {
	store := newTestStore(t)
	store.Save("only one", nil, testEmbedding)

	results, err := store.Search(testEmbedding, 100)

	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestStore_Search_PreservesTags(t *testing.T) {
	store := newTestStore(t)
	store.Save("tagged content", []string{"alpha", "beta"}, testEmbedding)

	results, err := store.Search(testEmbedding, 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, []string{"alpha", "beta"}, results[0].Tags)
}

// --------------- SearchFTS ---------------

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

func TestStore_SearchFTS_RespectsLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 10; i++ {
		store.Save(fmt.Sprintf("golang note %d", i), nil, nil)
	}

	results, err := store.SearchFTS("golang", 3)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestStore_SearchFTS_SetsScore(t *testing.T) {
	store := newTestStore(t)
	store.Save("buscando algo especial", nil, nil)

	results, err := store.SearchFTS("especial", 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Greater(t, results[0].Score, 0.0)
}

func TestStore_SearchFTS_EmptyStore(t *testing.T) {
	store := newTestStore(t)

	results, err := store.SearchFTS("anything", 5)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_SearchFTS_PreservesTags(t *testing.T) {
	store := newTestStore(t)
	store.Save("test content for fts", []string{"tag1", "tag2"}, nil)

	results, err := store.SearchFTS("test", 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, []string{"tag1", "tag2"}, results[0].Tags)
}

// --------------- SearchHybrid ---------------

func TestStore_SearchHybrid_CombinesResults(t *testing.T) {
	store := newTestStore(t)
	store.Save("el auto necesita nafta", []string{"transport"}, []float64{0.8, 0.1, 0.0, 0.0})
	store.Save("comprar leche en el super", []string{"food"}, []float64{0.0, 0.0, 0.8, 0.1})

	results, err := store.SearchHybrid("nafta auto", []float64{0.8, 0.1, 0.0, 0.0}, 5, 0.6, 0.4)

	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Contains(t, results[0].Content, "nafta")
}

func TestStore_SearchHybrid_EmptyStore(t *testing.T) {
	store := newTestStore(t)

	results, err := store.SearchHybrid("anything", testEmbedding, 5, 0.6, 0.4)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_SearchHybrid_RespectsLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 10; i++ {
		store.Save(fmt.Sprintf("golang note number %d", i), nil, testEmbedding)
	}

	results, err := store.SearchHybrid("golang", testEmbedding, 3, 0.6, 0.4)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 3)
}

// --------------- Conversations ---------------

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

func TestStore_LoadConversation_EmptySession(t *testing.T) {
	store := newTestStore(t)

	msgs, err := store.LoadConversation("nonexistent", 10)

	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestStore_LoadConversation_ReturnsChronologicalOrder(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "first")
	store.SaveConversation(testSessionID, "assistant", "second")
	store.SaveConversation(testSessionID, "user", "third")

	msgs, err := store.LoadConversation(testSessionID, 10)

	require.NoError(t, err)
	require.Len(t, msgs, 3)
	assert.Equal(t, "first", msgs[0].Content)
	assert.Equal(t, "second", msgs[1].Content)
	assert.Equal(t, "third", msgs[2].Content)
}

func TestStore_ClearConversation(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "msg")

	err := store.ClearConversation(testSessionID)

	require.NoError(t, err)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	assert.Empty(t, msgs)
}

func TestStore_ClearConversation_DoesNotAffectOtherSessions(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation("session-a", "user", "keep me")
	store.SaveConversation("session-b", "user", "delete me")

	store.ClearConversation("session-b")

	msgsA, _ := store.LoadConversation("session-a", 10)
	msgsB, _ := store.LoadConversation("session-b", 10)

	assert.Len(t, msgsA, 1)
	assert.Empty(t, msgsB)
}

func TestStore_ClearConversation_NonexistentSession_NoError(t *testing.T) {
	store := newTestStore(t)

	err := store.ClearConversation("nonexistent")

	assert.NoError(t, err)
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

func TestStore_ReplaceConversation_WithMultipleMessages(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "old")

	replacement := []domain.ConversationMessage{
		{Role: "user", Content: "summary question"},
		{Role: "assistant", Content: "summary answer"},
	}
	err := store.ReplaceConversation(testSessionID, replacement)

	require.NoError(t, err)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	require.Len(t, msgs, 2)
	assert.Equal(t, "summary question", msgs[0].Content)
	assert.Equal(t, "summary answer", msgs[1].Content)
}

func TestStore_ReplaceConversation_EmptySlice_ClearsSession(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "will be cleared")

	err := store.ReplaceConversation(testSessionID, []domain.ConversationMessage{})

	require.NoError(t, err)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	assert.Empty(t, msgs)
}

func TestStore_ReplaceConversation_DoesNotAffectOtherSessions(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation("session-a", "user", "keep me")
	store.SaveConversation("session-b", "user", "replace me")

	replacement := []domain.ConversationMessage{
		{Role: "assistant", Content: "replaced"},
	}
	store.ReplaceConversation("session-b", replacement)

	msgsA, _ := store.LoadConversation("session-a", 10)
	msgsB, _ := store.LoadConversation("session-b", 10)

	assert.Len(t, msgsA, 1)
	assert.Equal(t, "keep me", msgsA[0].Content)
	assert.Len(t, msgsB, 1)
	assert.Equal(t, "replaced", msgsB[0].Content)
}

// --------------- Habits ---------------

func TestStore_LogHabit_Success(t *testing.T) {
	store := newTestStore(t)

	err := store.LogHabit("exercise")

	assert.NoError(t, err)
}

func TestStore_LogHabit_MultipleTimes(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.LogHabit("exercise"))
	require.NoError(t, store.LogHabit("exercise"))
	require.NoError(t, store.LogHabit("meditation"))

	habits, err := store.ListHabitsToday()
	require.NoError(t, err)
	assert.Contains(t, habits, "exercise")
	assert.Contains(t, habits, "meditation")
}

func TestStore_ListHabitsToday_NoHabits(t *testing.T) {
	store := newTestStore(t)

	habits, err := store.ListHabitsToday()

	require.NoError(t, err)
	assert.Empty(t, habits)
}

func TestStore_ListHabitsToday_ReturnsDistinctNames(t *testing.T) {
	store := newTestStore(t)
	store.LogHabit("reading")
	store.LogHabit("reading")
	store.LogHabit("running")

	habits, err := store.ListHabitsToday()

	require.NoError(t, err)
	assert.Len(t, habits, 2)
}

func TestStore_GetHabitStreak_NoLogs(t *testing.T) {
	store := newTestStore(t)

	streak, total, err := store.GetHabitStreak("exercise")

	require.NoError(t, err)
	assert.Equal(t, 0, streak)
	assert.Equal(t, 0, total)
}

func TestStore_GetHabitStreak_TodayOnly(t *testing.T) {
	store := newTestStore(t)
	store.LogHabit("exercise")

	streak, total, err := store.GetHabitStreak("exercise")

	require.NoError(t, err)
	assert.Equal(t, 1, streak)
	assert.Equal(t, 1, total)
}

func TestStore_GetHabitStreak_TotalCountsAllLogs(t *testing.T) {
	store := newTestStore(t)
	store.LogHabit("exercise")
	store.LogHabit("exercise")

	_, total, err := store.GetHabitStreak("exercise")

	require.NoError(t, err)
	assert.Equal(t, 2, total)
}

func TestStore_GetHabitStreak_IsolatesHabitNames(t *testing.T) {
	store := newTestStore(t)
	store.LogHabit("exercise")
	store.LogHabit("meditation")

	streakEx, totalEx, err := store.GetHabitStreak("exercise")
	require.NoError(t, err)
	assert.Equal(t, 1, streakEx)
	assert.Equal(t, 1, totalEx)

	streakMed, totalMed, err := store.GetHabitStreak("meditation")
	require.NoError(t, err)
	assert.Equal(t, 1, streakMed)
	assert.Equal(t, 1, totalMed)
}

// --------------- ListExpenses ---------------

func TestStore_ListExpenses_EmptyTable(t *testing.T) {
	store := newTestStore(t)

	expenses, err := store.ListExpenses("2024-01-01", "2024-12-31")

	require.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestStore_ListExpenses_ReturnsExpenses(t *testing.T) {
	store := newTestStore(t)

	_, err := store.db.Exec(
		"INSERT INTO expenses (date, description, category, amount, amount_usd, paid_by) VALUES (?, ?, ?, ?, ?, ?)",
		"2024-06-15", "Supermercado Dia", "Supermercado", 5000.0, 5.5, "Sebas",
	)
	require.NoError(t, err)

	expenses, err := store.ListExpenses("2024-01-01", "2024-12-31")

	require.NoError(t, err)
	require.Len(t, expenses, 1)
	assert.Equal(t, "Supermercado Dia", expenses[0].Description)
	assert.Equal(t, "Supermercado", expenses[0].Category)
	assert.Equal(t, 5000.0, expenses[0].Amount)
	assert.Equal(t, 5.5, expenses[0].AmountUSD)
	assert.Equal(t, "Sebas", expenses[0].PaidBy)
}

func TestStore_ListExpenses_FiltersDateRange(t *testing.T) {
	store := newTestStore(t)

	store.db.Exec("INSERT INTO expenses (date, description, category, amount, amount_usd, paid_by) VALUES (?, ?, ?, ?, ?, ?)",
		"2024-01-15", "Enero", "Otro", 100.0, 0.1, "Sebas")
	store.db.Exec("INSERT INTO expenses (date, description, category, amount, amount_usd, paid_by) VALUES (?, ?, ?, ?, ?, ?)",
		"2024-06-15", "Junio", "Otro", 200.0, 0.2, "Sebas")
	store.db.Exec("INSERT INTO expenses (date, description, category, amount, amount_usd, paid_by) VALUES (?, ?, ?, ?, ?, ?)",
		"2024-12-15", "Diciembre", "Otro", 300.0, 0.3, "Sebas")

	expenses, err := store.ListExpenses("2024-05-01", "2024-07-31")

	require.NoError(t, err)
	require.Len(t, expenses, 1)
	assert.Equal(t, "Junio", expenses[0].Description)
}

func TestStore_ListExpenses_MultipleExpenses(t *testing.T) {
	store := newTestStore(t)

	for i := 0; i < 5; i++ {
		store.db.Exec("INSERT INTO expenses (date, description, category, amount, amount_usd, paid_by) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("2024-06-%02d", i+1), fmt.Sprintf("Expense %d", i), "Otro", float64(i*100), 0.0, "Sebas")
	}

	expenses, err := store.ListExpenses("2024-06-01", "2024-06-30")

	require.NoError(t, err)
	assert.Len(t, expenses, 5)
}

// --------------- PruneSessions ---------------

func TestStore_PruneSessions_NoOldSessions(t *testing.T) {
	store := newTestStore(t)
	store.SaveConversation(testSessionID, "user", "recent message")

	pruned, err := store.PruneSessions(7)

	require.NoError(t, err)
	assert.Equal(t, int64(0), pruned)

	msgs, _ := store.LoadConversation(testSessionID, 10)
	assert.Len(t, msgs, 1)
}

func TestStore_PruneSessions_EmptyTable(t *testing.T) {
	store := newTestStore(t)

	pruned, err := store.PruneSessions(7)

	require.NoError(t, err)
	assert.Equal(t, int64(0), pruned)
}

func TestStore_PruneSessions_OldSessionsPruned(t *testing.T) {
	store := newTestStore(t)

	_, err := store.db.Exec(
		"INSERT INTO conversations (session_id, role, content, created_at) VALUES (?, ?, ?, ?)",
		"old-session", "user", "old message", time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05"),
	)
	require.NoError(t, err)

	store.SaveConversation("recent-session", "user", "fresh message")

	pruned, err := store.PruneSessions(7)

	require.NoError(t, err)
	assert.Greater(t, pruned, int64(0))

	msgs, _ := store.LoadConversation("recent-session", 10)
	assert.Len(t, msgs, 1)

	oldMsgs, _ := store.LoadConversation("old-session", 10)
	assert.Empty(t, oldMsgs)
}

// --------------- Close ---------------

func TestStore_Close_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLiteMemoryService(dbPath)
	require.NoError(t, err)

	err = store.Close()

	assert.NoError(t, err)
}

func TestStore_Close_DoubleClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLiteMemoryService(dbPath)
	require.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	err = store.Close()
	assert.Error(t, err)
}
