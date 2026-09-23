package repository

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          int64
	Title       string
	Description sql.NullString
	Status      string
	Priority    int
	ProjectID   int64
	AssigneeID  sql.NullInt64
	DueDate     sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func GetAllTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`
		SELECT *
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
		SELECT * FROM tasks
		WHERE id = ?
	`, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil

}

func createTodo(db *sql.DB, task *Task) error {
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
		return nil
	}
	task.ID = id
	task.CreatedAt = now
	task.UpdatedAt = now

	return nil

}
