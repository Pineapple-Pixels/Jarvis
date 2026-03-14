package sqldata

import _ "embed"

// Postgres queries.

//go:embed postgres/insert/save_memory.sql
var PGSaveMemory string

//go:embed postgres/insert/save_conversation.sql
var PGSaveConversation string

//go:embed postgres/insert/log_habit.sql
var PGLogHabit string

//go:embed postgres/select/all_memories.sql
var PGSelectMemories string

//go:embed postgres/select/search_fts.sql
var PGSearchFTS string

//go:embed postgres/select/load_conversation.sql
var PGLoadConversation string

//go:embed postgres/select/count_habits.sql
var PGCountHabit string

//go:embed postgres/select/habit_dates.sql
var PGHabitDates string

//go:embed postgres/select/habits_today.sql
var PGHabitsToday string

//go:embed postgres/select/list_expenses.sql
var PGListExpenses string

//go:embed postgres/delete/delete_memory.sql
var PGDeleteMemory string

//go:embed postgres/delete/clear_conversation.sql
var PGClearConversation string
