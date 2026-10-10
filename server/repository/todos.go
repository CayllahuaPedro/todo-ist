package repository

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description sql.NullString `json:"description"`
	Status      string         `json:"status"`
	Priority    int            `json:"priority"`
	ProjectID   int64          `json:"project_id"`
	AssigneeID  sql.NullInt64  `json:"assignee_id"`
	DueDate     sql.NullTime   `json:"due_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func GetAllTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := []Task{}

	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func GetTaskById(db *sql.DB, id int) (*Task, error) {
	var task Task
	err := db.QueryRow(`
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func CreateTodo(db *sql.DB, task *Task) error {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, task.Title, task.Description, task.Status, now, now)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = id
	task.CreatedAt = now
	task.UpdatedAt = now

	return nil

}

func UpdateTodo(db *sql.DB, task *Task) error {
	now := time.Now()
	result, err := db.Exec(`
		UPDATE tasks SET
		title = ?,
		description = ?,
		status = ?,
		updated_at = ?
		WHERE id = ?
	`, task.Title, task.Description, task.Status, now, task.ID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func DeleteTodo(db *sql.DB, id int) error {
	result, err := db.Exec(`
		DELETE FROM tasks WHERE id = ?
	`, id)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func requireAffected(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
