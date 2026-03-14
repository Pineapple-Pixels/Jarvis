package domain

// Jira path parameter constants.
const (
	PathParamKey = "key"
)

// JiraIssue represents a Jira issue.
type JiraIssue struct {
	Key      string `json:"key"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
	URL      string `json:"url"`
	Type     string `json:"type"`
}

// JiraCreateIssueRequest is the payload for creating a Jira issue.
type JiraCreateIssueRequest struct {
	ProjectKey  string `json:"project_key"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	IssueType   string `json:"issue_type"`
}

// JiraIssueListResponse is the response for listing issues.
type JiraIssueListResponse struct {
	Success bool        `json:"success"`
	Issues  []JiraIssue `json:"issues,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JiraIssueResponse is the response for a single issue operation.
type JiraIssueResponse struct {
	Success bool       `json:"success"`
	Issue   *JiraIssue `json:"issue,omitempty"`
	Error   string     `json:"error,omitempty"`
}
