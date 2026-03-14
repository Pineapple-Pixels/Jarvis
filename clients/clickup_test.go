package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClickUpServer(t *testing.T, handlers map[string]http.HandlerFunc) (*httptest.Server, *ClickUpClient) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := handlers[r.Method+" "+r.URL.Path]; ok {
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	client := NewClickUpClient("test-token", "team123")
	client.httpClient = srv.Client()

	// Override base URL by replacing doRequest
	origBaseURL := clickupBaseURL
	t.Cleanup(func() {
		srv.Close()
		_ = origBaseURL // keep reference
	})

	// We need to point the client at the test server. Since clickupBaseURL is a const,
	// we'll wrap the server to intercept at the transport level.
	client.httpClient.Transport = &rewriteTransport{
		base:    srv.Client().Transport,
		baseURL: srv.URL,
	}

	return srv, client
}

type rewriteTransport struct {
	base    http.RoundTripper
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = t.baseURL[len("http://"):]
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestClickUpClient_GetMyTasks_Success(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GET /api/v2/user": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 12345},
			})
		},
		"GET /api/v2/team/team123/task": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "12345", r.URL.Query().Get("assignees[]"))
			json.NewEncoder(w).Encode(map[string]any{
				"tasks": []map[string]any{
					{
						"id":   "t1",
						"name": "Fix bug",
						"url":  "https://app.clickup.com/t/t1",
						"status": map[string]string{
							"status": "open",
						},
						"assignees": []map[string]string{
							{"username": "john"},
						},
						"due_date": "1679875200000",
					},
				},
			})
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	tasks, err := client.GetMyTasks()

	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "t1", tasks[0].ID)
	assert.Equal(t, "Fix bug", tasks[0].Name)
	assert.Equal(t, "open", tasks[0].Status)
	assert.Equal(t, []string{"john"}, tasks[0].Assignees)
}

func TestClickUpClient_GetMyTasks_UserError(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GET /api/v2/user": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"err":"Token invalid"}`))
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	tasks, err := client.GetMyTasks()

	assert.Error(t, err)
	assert.Nil(t, tasks)
	assert.Contains(t, err.Error(), "401")
}

func TestClickUpClient_GetTask_Success(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GET /api/v2/task/abc": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{
				"id":   "abc",
				"name": "Deploy",
				"url":  "https://app.clickup.com/t/abc",
				"status": map[string]string{
					"status": "in progress",
				},
				"assignees": []map[string]string{},
				"due_date":  "",
			})
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	task, err := client.GetTask("abc")

	require.NoError(t, err)
	assert.Equal(t, "abc", task.ID)
	assert.Equal(t, "Deploy", task.Name)
	assert.Equal(t, "in progress", task.Status)
}

func TestClickUpClient_GetTask_APIError(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GET /api/v2/task/notfound": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"err":"Task not found"}`))
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	_, err := client.GetTask("notfound")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestClickUpClient_CreateTask_Success(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"POST /api/v2/list/lst1/task": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "New task", body["name"])
			assert.Equal(t, "desc", body["description"])

			json.NewEncoder(w).Encode(map[string]any{
				"id":   "new1",
				"name": "New task",
				"url":  "https://app.clickup.com/t/new1",
				"status": map[string]string{
					"status": "to do",
				},
				"assignees": []map[string]string{
					{"username": "me"},
				},
				"due_date": "",
			})
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	task, err := client.CreateTask("lst1", "New task", "desc")

	require.NoError(t, err)
	assert.Equal(t, "new1", task.ID)
	assert.Equal(t, "New task", task.Name)
	assert.Equal(t, "to do", task.Status)
	assert.Equal(t, []string{"me"}, task.Assignees)
}

func TestClickUpClient_UpdateTaskStatus_Success(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"PUT /api/v2/task/t1": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "done", body["status"])

			json.NewEncoder(w).Encode(map[string]any{
				"id": "t1", "name": "task", "status": map[string]string{"status": "done"},
			})
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	err := client.UpdateTaskStatus("t1", "done")

	assert.NoError(t, err)
}

func TestClickUpClient_UpdateTaskStatus_APIError(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"PUT /api/v2/task/t1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"err":"Forbidden"}`))
		},
	}

	_, client := newTestClickUpServer(t, handlers)

	err := client.UpdateTaskStatus("t1", "done")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestClickUpClient_AuthHeader(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"GET /api/v2/task/x": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "my-secret-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			json.NewEncoder(w).Encode(map[string]any{
				"id": "x", "name": "t", "url": "", "status": map[string]string{"status": "open"},
				"assignees": []any{}, "due_date": "",
			})
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := handlers[r.Method+" "+r.URL.Path]; ok {
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClickUpClient("my-secret-token", "team1")
	client.httpClient = &http.Client{
		Transport: &rewriteTransport{baseURL: srv.URL},
	}

	_, err := client.GetTask("x")

	assert.NoError(t, err)
}
