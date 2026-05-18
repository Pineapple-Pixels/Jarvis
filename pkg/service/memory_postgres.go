package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	appdb "jarvis/db"
	"jarvis/pkg/domain"
	"jarvis/pkg/service/sqldata"

	_ "github.com/lib/pq"
)

// scanRows iterates rows, applies scanner to each row, and collects the results.
// It handles rows.Close (via the caller's defer) and rows.Err automatically.
// T can be any type; the scanner function owns all Scan + Unmarshal calls.
func scanRows[T any](rows *sql.Rows, scanner func(*sql.Rows) (T, error)) ([]T, error) {
	var results []T
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			// Log and skip the bad row — consistent with the previous behaviour.
			log.Printf("pgstore: scan row: %v", err)
			continue
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return results, fmt.Errorf("pgstore: rows iteration: %w", err)
	}
	return results, nil
}

type PGMemoryService struct {
	db *sql.DB
}

func NewPGMemoryService(dsn string) (*PGMemoryService, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreOpen, err)
	}

	if err := db.Ping(); err != nil {
		return nil, domain.Wrapf(domain.ErrStorePing, err)
	}

	if err := appdb.RunMigrations(db); err != nil {
		return nil, domain.Wrapf(domain.ErrStoreMigrate, err)
	}

	return &PGMemoryService{db: db}, nil
}

func (s *PGMemoryService) Save(content string, tags []string, embedding []float64) (int64, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	embJSON, err := json.Marshal(embedding)
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	var id int64
	if err := s.db.QueryRow(sqldata.SaveMemory, content, string(tagsJSON), string(embJSON)).Scan(&id); err != nil {
		return 0, domain.Wrapf(domain.ErrStoreSave, err)
	}

	return id, nil
}

type scoredMemory struct {
	mem   domain.Memory
	score float64
}

func (s *PGMemoryService) Search(queryEmbedding []float64, limit int) ([]domain.Memory, error) {
	// candidateMultiplier gives the SQL query a wider pool to rank from while
	// still capping the table scan. Cosine similarity + time-decay reranks
	// within this candidate set in application code (embeddings are JSONB, not
	// pgvector, so in-DB distance is not available yet).
	const candidateMultiplier = 10
	sqlLimit := limit * candidateMultiplier

	rows, err := s.db.Query(sqldata.SelectMemories, sqlLimit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreSearch, err)
	}
	defer rows.Close()

	scored, err := scanRows(rows, func(r *sql.Rows) (scoredMemory, error) {
		var m domain.Memory
		var tagsStr, embStr string
		if err := r.Scan(&m.ID, &m.Content, &tagsStr, &embStr, &m.CreatedAt); err != nil {
			return scoredMemory{}, fmt.Errorf("scan memory id: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsStr), &m.Tags); err != nil {
			log.Printf("pgstore: unmarshal tags id=%d: %v", m.ID, err)
		}
		var emb []float64
		if err := json.Unmarshal([]byte(embStr), &emb); err != nil {
			log.Printf("pgstore: unmarshal embedding id=%d: %v", m.ID, err)
		}
		similarity := cosineSimilarity(queryEmbedding, emb)
		decay := timeDecay(m.CreatedAt)
		return scoredMemory{mem: m, score: similarity * decay}, nil
	})
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreSearch, err)
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if limit > len(scored) {
		limit = len(scored)
	}

	memories := make([]domain.Memory, limit)
	for i := 0; i < limit; i++ {
		memories[i] = scored[i].mem
		memories[i].Score = scored[i].score
	}

	return memories, nil
}

func (s *PGMemoryService) SearchFTS(query string, limit int) ([]domain.Memory, error) {
	rows, err := s.db.Query(sqldata.SearchFTS, query, limit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreFTS, err)
	}
	defer rows.Close()

	results, err := scanRows(rows, func(r *sql.Rows) (domain.Memory, error) {
		var m domain.Memory
		var tagsStr string
		var rank float64
		if err := r.Scan(&m.ID, &m.Content, &tagsStr, &m.CreatedAt, &rank); err != nil {
			return domain.Memory{}, fmt.Errorf("scan fts row: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsStr), &m.Tags); err != nil {
			log.Printf("pgstore: unmarshal tags id=%d: %v", m.ID, err)
		}
		m.Score = rank * timeDecay(m.CreatedAt)
		return m, nil
	})
	if err != nil {
		return nil, domain.Wrapf(domain.ErrStoreFTS, err)
	}

	return results, nil
}

func (s *PGMemoryService) SearchHybrid(query string, queryEmbedding []float64, limit int, vecWeight, ftsWeight float64) ([]domain.Memory, error) {
	vecResults, vecErr := s.Search(queryEmbedding, limit*2)
	ftsResults, ftsErr := s.SearchFTS(query, limit*2)

	if vecErr != nil && ftsErr != nil {
		return nil, domain.Wrap(domain.ErrStoreHybrid, "both vector and fts failed")
	}

	merged := mergeSearchResults(vecResults, ftsResults, vecErr, ftsErr, vecWeight, ftsWeight)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	if limit > len(merged) {
		limit = len(merged)
	}

	return merged[:limit], nil
}

func (s *PGMemoryService) Delete(id int64) error {
	if _, err := s.db.Exec(sqldata.DeleteMemory, id); err != nil {
		return domain.Wrapf(domain.ErrStoreDelete, err)
	}
	return nil
}

func (s *PGMemoryService) SaveConversation(sessionID, role, content string) error {
	if _, err := s.db.Exec(sqldata.SaveConversation, sessionID, role, content); err != nil {
		return domain.Wrapf(domain.ErrConversationSave, err)
	}
	return nil
}

func (s *PGMemoryService) LoadConversation(sessionID string, limit int) ([]domain.ConversationMessage, error) {
	rows, err := s.db.Query(sqldata.LoadConversation, sessionID, limit)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrConversationLoad, err)
	}
	defer rows.Close()

	msgs, err := scanRows(rows, func(r *sql.Rows) (domain.ConversationMessage, error) {
		var m domain.ConversationMessage
		if err := r.Scan(&m.Role, &m.Content, &m.CreatedAt); err != nil {
			return domain.ConversationMessage{}, fmt.Errorf("scan conversation row: %w", err)
		}
		return m, nil
	})
	if err != nil {
		return nil, domain.Wrapf(domain.ErrConversationLoad, err)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}

func (s *PGMemoryService) ClearConversation(sessionID string) error {
	if _, err := s.db.Exec(sqldata.ClearConversation, sessionID); err != nil {
		return domain.Wrapf(domain.ErrConversationClear, err)
	}
	return nil
}

func (s *PGMemoryService) ReplaceConversation(sessionID string, msgs []domain.ConversationMessage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return domain.Wrapf(domain.ErrConversationReplace, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sqldata.ClearConversation, sessionID); err != nil {
		return domain.Wrapf(domain.ErrConversationReplace, err)
	}

	for _, m := range msgs {
		if _, err := tx.Exec(sqldata.SaveConversation, sessionID, m.Role, m.Content); err != nil {
			return domain.Wrapf(domain.ErrConversationReplace, err)
		}
	}

	return tx.Commit()
}

func (s *PGMemoryService) LogHabit(name string) error {
	if _, err := s.db.Exec(sqldata.LogHabit, name); err != nil {
		return domain.Wrapf(domain.ErrHabitLog, err)
	}
	return nil
}

func (s *PGMemoryService) GetHabitStreak(name string) (int, int, error) {
	var total int
	if err := s.db.QueryRow(sqldata.CountHabit, name).Scan(&total); err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}

	rows, err := s.db.Query(sqldata.HabitDates, name)
	if err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	defer rows.Close()

	dates, err := scanRows(rows, func(r *sql.Rows) (time.Time, error) {
		var d time.Time
		if err := r.Scan(&d); err != nil {
			return time.Time{}, fmt.Errorf("scan habit date: %w", err)
		}
		return d, nil
	})
	if err != nil {
		return 0, 0, domain.Wrapf(domain.ErrHabitQuery, err)
	}

	streak := 0
	expected := time.Now().Truncate(24 * time.Hour)
	for _, d := range dates {
		d = d.Truncate(24 * time.Hour)
		if d.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, total, nil
}

func (s *PGMemoryService) ListHabitsToday() ([]string, error) {
	rows, err := s.db.Query(sqldata.HabitsToday)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrHabitQuery, err)
	}
	defer rows.Close()

	habits, err := scanRows(rows, func(r *sql.Rows) (string, error) {
		var name string
		if err := r.Scan(&name); err != nil {
			return "", fmt.Errorf("scan habit row: %w", err)
		}
		return name, nil
	})
	if err != nil {
		return nil, domain.Wrapf(domain.ErrHabitQuery, err)
	}

	return habits, nil
}

func (s *PGMemoryService) ListExpenses(from, to string) ([]domain.Expense, error) {
	rows, err := s.db.Query(sqldata.ListExpenses, from, to)
	if err != nil {
		return nil, domain.Wrapf(domain.ErrFinanceSummary, err)
	}
	defer rows.Close()

	expenses, err := scanRows(rows, func(r *sql.Rows) (domain.Expense, error) {
		var e domain.Expense
		if err := r.Scan(&e.ID, &e.Date, &e.Description, &e.Category, &e.Amount, &e.AmountUSD, &e.PaidBy); err != nil {
			return domain.Expense{}, fmt.Errorf("scan expense row: %w", err)
		}
		return e, nil
	})
	if err != nil {
		return nil, domain.Wrapf(domain.ErrFinanceSummary, err)
	}

	return expenses, nil
}

func (s *PGMemoryService) PruneSessions(olderThanDays int) (int64, error) {
	result, err := s.db.Exec(sqldata.PruneConversations, fmt.Sprintf("%d", olderThanDays))
	if err != nil {
		return 0, domain.Wrapf(domain.ErrStoreDelete, err)
	}
	return result.RowsAffected()
}

func (s *PGMemoryService) Close() error {
	return s.db.Close()
}

var _ MemoryService = (*PGMemoryService)(nil)
