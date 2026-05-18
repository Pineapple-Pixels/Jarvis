---
name: project_test_baseline
description: Jarvis test suite baseline analysis — coverage gaps and quality issues found in May 2026 audit
metadata:
  type: project
---

Baseline audit of the Jarvis Go test suite conducted 2026-05-17.

Key facts:
- 729 test functions across 70 test files
- No integration tests hitting real Postgres — all service layer is mocked
- 3 controllers have zero test coverage: catalog, health, skills_qa

**Why:** First full QA audit of the project. Gaps are the prioritized backlog.
**How to apply:** Use this as the baseline when tracking coverage improvements over time. P1 items should be addressed before any production release.

Priority gaps (P1):
- pkg/controller/catalog.go, health.go, skills_qa.go — zero tests
- pkg/service/memory_postgres.go — 16 exported methods, zero integration tests

Priority gaps (P2):
- pkg/controller/gmail_test.go — only tests GetMessage, missing ListUnread
- pkg/controller/calendar_test.go — only tests validation errors, no success path
- internal/middleware/trace.go — no test file at all
- domain.TypingIndicator interface — not mocked, not tested
- domain.FailoverProvider — no unit test coverage

Priority gaps (P3):
- pkg/usecase/ratelimit_test.go — no concurrent access test (race condition risk)
- pkg/usecase/orchestrator_test.go — missing Run() error paths (tool execution failure)
- MockCatalogService missing from test/mocks.go (only NullCatalogService tested)
