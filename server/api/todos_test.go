package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todo-ist/api"
	"todo-ist/database"
)

type taskResponse struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mux := http.NewServeMux()
	api.DBHandler{DB: db}.RegisterRoutes(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func decode[T any](t *testing.T, res *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return v
}

func readBody(t *testing.T, res *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return strings.TrimSpace(string(b))
}

func assertStatus(t *testing.T, res *http.Response, want int) {
	t.Helper()
	if res.StatusCode != want {
		t.Fatalf("%s %s: status = %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
}

func createTask(t *testing.T, srv *httptest.Server, title string) taskResponse {
	t.Helper()
	res := do(t, srv, http.MethodPost, "/todos", fmt.Sprintf(`{"title":%q,"status":"pending"}`, title))
	assertStatus(t, res, http.StatusCreated)
	return decode[taskResponse](t, res)
}

func TestListTodosStartsEmpty(t *testing.T) {
	srv := newServer(t)

	res := do(t, srv, http.MethodGet, "/todos", "")

	assertStatus(t, res, http.StatusOK)
	if tasks := decode[[]taskResponse](t, res); len(tasks) != 0 {
		t.Fatalf("got %d tasks, want 0", len(tasks))
	}
}

func TestCreateTodo(t *testing.T) {
	srv := newServer(t)

	created := createTask(t, srv, "Buy coffee")

	if created.ID == 0 || created.Title != "Buy coffee" || created.Status != "pending" {
		t.Fatalf("unexpected created task: %+v", created)
	}
	tasks := decode[[]taskResponse](t, do(t, srv, http.MethodGet, "/todos", ""))
	if len(tasks) != 1 || tasks[0].ID != created.ID {
		t.Fatalf("list = %+v, want only task %d", tasks, created.ID)
	}
}

func TestCreateTodoRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		message string
	}{
		{"malformed json", `{"title":`, "invalid JSON"},
		{"missing title", `{"title":"","status":"pending"}`, "title is required"},
		{"unknown status", `{"title":"Buy coffee","status":"later"}`, "invalid status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t)

			res := do(t, srv, http.MethodPost, "/todos", tc.body)

			assertStatus(t, res, http.StatusBadRequest)
			if got := readBody(t, res); got != tc.message {
				t.Fatalf("message = %q, want %q", got, tc.message)
			}
		})
	}
}

func TestGetTodo(t *testing.T) {
	srv := newServer(t)
	created := createTask(t, srv, "Buy coffee")

	res := do(t, srv, http.MethodGet, fmt.Sprintf("/todos/%d", created.ID), "")

	assertStatus(t, res, http.StatusOK)
	if got := decode[taskResponse](t, res); got != created {
		t.Fatalf("got %+v, want %+v", got, created)
	}
}

func TestGetTodoErrors(t *testing.T) {
	srv := newServer(t)

	assertStatus(t, do(t, srv, http.MethodGet, "/todos/999", ""), http.StatusNotFound)
	assertStatus(t, do(t, srv, http.MethodGet, "/todos/abc", ""), http.StatusBadRequest)
}

func TestUpdateTodo(t *testing.T) {
	srv := newServer(t)
	created := createTask(t, srv, "Buy coffee")
	path := fmt.Sprintf("/todos/%d", created.ID)

	res := do(t, srv, http.MethodPatch, path, `{"title":"Buy beans","status":"completed"}`)

	assertStatus(t, res, http.StatusNoContent)
	got := decode[taskResponse](t, do(t, srv, http.MethodGet, path, ""))
	if got.Title != "Buy beans" || got.Status != "completed" {
		t.Fatalf("after update got %+v", got)
	}
}

func TestUpdateTodoErrors(t *testing.T) {
	srv := newServer(t)
	valid := `{"title":"Buy coffee","status":"pending"}`

	assertStatus(t, do(t, srv, http.MethodPatch, "/todos/999", valid), http.StatusNotFound)
	assertStatus(t, do(t, srv, http.MethodPatch, "/todos/abc", valid), http.StatusBadRequest)
	assertStatus(t, do(t, srv, http.MethodPatch, "/todos/1", `{"title":`), http.StatusBadRequest)
}

func TestDeleteTodo(t *testing.T) {
	srv := newServer(t)
	created := createTask(t, srv, "Buy coffee")
	path := fmt.Sprintf("/todos/%d", created.ID)

	assertStatus(t, do(t, srv, http.MethodDelete, path, ""), http.StatusNoContent)
	assertStatus(t, do(t, srv, http.MethodGet, path, ""), http.StatusNotFound)
}

func TestDeleteTodoErrors(t *testing.T) {
	srv := newServer(t)

	assertStatus(t, do(t, srv, http.MethodDelete, "/todos/999", ""), http.StatusNotFound)
	assertStatus(t, do(t, srv, http.MethodDelete, "/todos/abc", ""), http.StatusBadRequest)
}
