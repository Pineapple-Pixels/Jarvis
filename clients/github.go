package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"jarvis/pkg/domain"
)

const (
	githubBaseURL        = "https://api.github.com"
	githubDefaultTimeout = 15 * time.Second
	githubAcceptJSON     = "application/vnd.github+json"
)

// githubRepo is the raw GitHub API response shape for a repository.
type githubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"html_url"`
}

// githubIssue is the raw GitHub API response shape for an issue.
type githubIssue struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
}

// githubPullRequest is the raw GitHub API response shape for a pull request.
type githubPullRequest struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
	Draft  bool   `json:"draft"`
}

// GitHubClient is the GitHub API client.
type GitHubClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		token:      token,
		baseURL:    githubBaseURL,
		httpClient: &http.Client{Timeout: githubDefaultTimeout},
	}
}

// NewGitHubClientWithBaseURL creates a GitHub client pointing at a custom base URL (for testing).
func NewGitHubClientWithBaseURL(token, baseURL string) *GitHubClient {
	c := NewGitHubClient(token)
	c.baseURL = baseURL
	return c
}

// ListRepos returns the authenticated user's repositories.
func (c *GitHubClient) ListRepos() ([]domain.GitHubRepo, error) {
	resp, err := c.doRequest(http.MethodGet, "/user/repos", nil)
	if err != nil {
		return nil, err
	}

	var raw []githubRepo
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, fmt.Errorf("github: parse repos: %w", err)
	}

	repos := make([]domain.GitHubRepo, len(raw))
	for i, r := range raw {
		repos[i] = domain.GitHubRepo{ID: r.ID, Name: r.Name, FullName: r.FullName, URL: r.URL}
	}
	return repos, nil
}

// ListIssues returns issues for the given repository.
func (c *GitHubClient) ListIssues(owner, repo string) ([]domain.GitHubIssue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var raw []githubIssue
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, fmt.Errorf("github: parse issues: %w", err)
	}

	issues := make([]domain.GitHubIssue, len(raw))
	for i, iss := range raw {
		issues[i] = domain.GitHubIssue{ID: iss.ID, Number: iss.Number, Title: iss.Title, State: iss.State, URL: iss.URL}
	}
	return issues, nil
}

// CreateIssue creates a new issue in the given repository.
func (c *GitHubClient) CreateIssue(owner, repo, title, body string) (domain.GitHubIssue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	resp, err := c.doRequest(http.MethodPost, path, payload)
	if err != nil {
		return domain.GitHubIssue{}, err
	}

	var raw githubIssue
	if err := json.Unmarshal(resp, &raw); err != nil {
		return domain.GitHubIssue{}, fmt.Errorf("github: parse create issue: %w", err)
	}

	return domain.GitHubIssue{ID: raw.ID, Number: raw.Number, Title: raw.Title, State: raw.State, URL: raw.URL}, nil
}

// ListPRs returns pull requests for the given repository.
func (c *GitHubClient) ListPRs(owner, repo string) ([]domain.GitHubPullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var raw []githubPullRequest
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, fmt.Errorf("github: parse pull requests: %w", err)
	}

	prs := make([]domain.GitHubPullRequest, len(raw))
	for i, pr := range raw {
		prs[i] = domain.GitHubPullRequest{ID: pr.ID, Number: pr.Number, Title: pr.Title, State: pr.State, URL: pr.URL, Draft: pr.Draft}
	}
	return prs, nil
}

func (c *GitHubClient) doRequest(method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("github: marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}

	req.Header.Set(headerAuthorization, "Bearer "+c.token)
	req.Header.Set("Accept", githubAcceptJSON)
	req.Header.Set(headerContentType, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("github: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
