package sqldata

import _ "embed"

// SQLite queries.

//go:embed sqlite/insert/save_memory.sql
var SaveMemory string

//go:embed sqlite/insert/save_conversation.sql
var SaveConversation string

//go:embed sqlite/insert/log_habit.sql
var LogHabit string

//go:embed sqlite/select/all_memories.sql
var SelectMemories string

//go:embed sqlite/select/search_fts.sql
var SearchFTS string

//go:embed sqlite/select/load_conversation.sql
var LoadConversation string

//go:embed sqlite/select/count_habits.sql
var CountHabit string

//go:embed sqlite/select/habit_dates.sql
var HabitDates string

//go:embed sqlite/select/habits_today.sql
var HabitsToday string

//go:embed sqlite/select/list_expenses.sql
var ListExpenses string

//go:embed sqlite/delete/delete_memory.sql
var DeleteMemory string

//go:embed sqlite/delete/clear_conversation.sql
var ClearConversation string

//go:embed sqlite/delete/prune_conversations.sql
var PruneConversations string
