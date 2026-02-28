package sqlite

import (
	"database/sql"
	"time"

	"lingosql/internal/models"
)

type TaskDAO struct {
	db *sql.DB
}

func NewTaskDAO(db *sql.DB) *TaskDAO {
	return &TaskDAO{db: db}
}

func (dao *TaskDAO) Create(task *models.Task) error {
	query := `INSERT INTO tasks (user_id, type, status, progress, payload, result, error_message, created_at, updated_at, started_at, finished_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := dao.db.Exec(
		query,
		task.UserID,
		task.Type,
		task.Status,
		task.Progress,
		task.Payload,
		task.Result,
		task.ErrorMessage,
		time.Now(),
		time.Now(),
		task.StartedAt,
		task.FinishedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)
	return nil
}

func (dao *TaskDAO) UpdateStatus(taskID int, status string, progress int, result string, errorMessage string, startedAt *time.Time, finishedAt *time.Time) error {
	query := `UPDATE tasks 
	          SET status = ?, progress = ?, result = ?, error_message = ?, 
	              started_at = COALESCE(?, started_at), finished_at = ?, updated_at = ? 
	          WHERE id = ?`
	_, err := dao.db.Exec(query, status, progress, result, errorMessage, startedAt, finishedAt, time.Now(), taskID)
	return err
}

func (dao *TaskDAO) Cancel(taskID int, finishedAt *time.Time) error {
	query := `UPDATE tasks SET status = ?, progress = ?, finished_at = ?, updated_at = ? WHERE id = ?`
	_, err := dao.db.Exec(query, "canceled", 100, finishedAt, time.Now(), taskID)
	return err
}

func (dao *TaskDAO) UpdateProgress(taskID int, progress int) error {
	query := `UPDATE tasks SET progress = ?, updated_at = ? WHERE id = ?`
	_, err := dao.db.Exec(query, progress, time.Now(), taskID)
	return err
}

func (dao *TaskDAO) GetByID(taskID int) (*models.Task, error) {
	task := &models.Task{}
	query := `SELECT id, user_id, type, status, progress, payload, result, error_message, created_at, updated_at, started_at, finished_at FROM tasks WHERE id = ?`
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	err := dao.db.QueryRow(query, taskID).Scan(
		&task.ID, &task.UserID, &task.Type, &task.Status, &task.Progress,
		&task.Payload, &task.Result, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt, &startedAt, &finishedAt,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		task.FinishedAt = &finishedAt.Time
	}
	return task, nil
}

func (dao *TaskDAO) ListByUser(userID int, page, pageSize int) ([]models.Task, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	if err := dao.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, type, status, progress, payload, result, error_message, created_at, updated_at, started_at, finished_at
	          FROM tasks WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := dao.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Type, &task.Status, &task.Progress,
			&task.Payload, &task.Result, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt, &startedAt, &finishedAt,
		); err != nil {
			continue
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			task.FinishedAt = &finishedAt.Time
		}
		tasks = append(tasks, task)
	}
	return tasks, total, nil
}
