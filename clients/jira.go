package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	jiraDefaultTimeout = 15 * time.Second

	jiraSearchPath  = "/rest/api/3/search"
	jiraIssuePath   = "/rest/api/3/issue"
	jiraMyIssuesJQL = "assignee=currentuser()"
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

// JiraClient is the Jira API client.
type JiraClient struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// NewJiraClient creates a new Jira API client.
func NewJiraClient(baseURL, email, apiToken string) *JiraClient {
	return &JiraClient{
		baseURL:    baseURL,
		email:      email,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: jiraDefaultTimeout},
	}
}

// GetMyIssues returns issues assigned to the current user.
func (c *JiraClient) GetMyIssues() ([]JiraIssue, error) {
	resp, err := c.doRequest(http.MethodGet, jiraSearchPath+"?jql="+jiraMyIssuesJQL, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Issues []struct {
			Key    string `json:"key"`
			Self   string `json:"self"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
				Assignee struct {
					DisplayName string `json:"displayName"`
				} `json:"assignee"`
				IssueType struct {
					Name string `json:"name"`
				} `json:"issuetype"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("jira: parse search response: %w", err)
	}

	issues := make([]JiraIssue, len(result.Issues))
	for i, raw := range result.Issues {
		issues[i] = JiraIssue{
			Key:      raw.Key,
			Summary:  raw.Fields.Summary,
			Status:   raw.Fields.Status.Name,
			Assignee: raw.Fields.Assignee.DisplayName,
			URL:      c.baseURL + "/browse/" + raw.Key,
			Type:     raw.Fields.IssueType.Name,
		}
	}

	return issues, nil
}

// GetIssue returns a single issue by key.
func (c *JiraClient) GetIssue(issueKey string) (JiraIssue, error) {
	resp, err := c.doRequest(http.MethodGet, jiraIssuePath+"/"+issueKey, nil)
	if err != nil {
		return JiraIssue{}, err
	}

	var raw struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
			Status  struct {
				Name string `json:"name"`
			} `json:"status"`
			Assignee struct {
				DisplayName string `json:"displayName"`
			} `json:"assignee"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issuetype"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(resp, &raw); err != nil {
		return JiraIssue{}, fmt.Errorf("jira: parse issue: %w", err)
	}

	return JiraIssue{
		Key:      raw.Key,
		Summary:  raw.Fields.Summary,
		Status:   raw.Fields.Status.Name,
		Assignee: raw.Fields.Assignee.DisplayName,
		URL:      c.baseURL + "/browse/" + raw.Key,
		Type:     raw.Fields.IssueType.Name,
	}, nil
}

// CreateIssue creates a new Jira issue.
func (c *JiraClient) CreateIssue(projectKey, summary, description, issueType string) (JiraIssue, error) {
	body := map[string]any{
		"fields": map[string]any{
			"project":   map[string]string{"key": projectKey},
			"summary":   summary,
			"description": map[string]any{
				"type":    "doc",
				"version": 1,
				"content": []map[string]any{
					{
						"type": "paragraph",
						"content": []map[string]any{
							{"type": "text", "text": description},
						},
					},
				},
			},
			"issuetype": map[string]string{"name": issueType},
		},
	}

	resp, err := c.doRequest(http.MethodPost, jiraIssuePath, body)
	if err != nil {
		return JiraIssue{}, err
	}

	var result struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return JiraIssue{}, fmt.Errorf("jira: parse create response: %w", err)
	}

	return c.GetIssue(result.Key)
}

// TransitionIssue transitions an issue to a new status.
func (c *JiraClient) TransitionIssue(issueKey, transitionID string) error {
	body := map[string]any{
		"transition": map[string]string{"id": transitionID},
	}

	_, err := c.doRequest(http.MethodPost, jiraIssuePath+"/"+issueKey+"/transitions", body)
	return err
}

func (c *JiraClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("jira: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("jira: create request: %w", err)
	}

	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jira: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("jira: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
