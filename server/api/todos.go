package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"todo-ist/repository"
)

type DBHandler struct {
	DB *sql.DB
}

type Status string

const (
	Pending    Status = "pending"
	InProgress Status = "in_progress"
	Cancelled  Status = "cancelled"
	Completed  Status = "completed"
)

func (status Status) IsValid() bool {
	switch status {
	case Pending, InProgress, Cancelled, Completed:
		return true
	}
	return false
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

type updateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (req createTaskRequest) Validate() error {
	if req.Title == "" {
		return errors.New("email is required")
	}
	if !req.Status.IsValid() {
		return errors.New("el estado es invalido")
	}
	return nil
}

func (db DBHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	tasks, err := repository.GetAllTasks(db.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (db DBHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	task, err := repository.GetTaskById(db.DB, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (db DBHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task := repository.Task{
		Title: req.Title,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Status: string(req.Status),
	}
	if err := repository.CreateTodo(db.DB, &task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (db DBHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task := repository.Task{
		ID:    id,
		Title: req.Title,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Status: string(req.Status),
	}
	if err := repository.UpdateTodo(db.DB, &task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (db DBHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := repository.DeleteTodo(db.DB, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
