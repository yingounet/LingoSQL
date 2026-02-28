package service

import (
	"encoding/json"
	"errors"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
)

type TaskService struct {
	taskDAO *sqlite.TaskDAO
}

func NewTaskService(taskDAO *sqlite.TaskDAO) *TaskService {
	return &TaskService{taskDAO: taskDAO}
}

func (s *TaskService) Create(userID int, taskType string, payload interface{}) (*models.Task, error) {
	payloadJSON := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			payloadJSON = string(b)
		}
	}
	task := &models.Task{
		UserID:   userID,
		Type:     taskType,
		Status:   "pending",
		Progress: 0,
		Payload:  payloadJSON,
	}
	if err := s.taskDAO.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Start(taskID int) error {
	now := time.Now()
	return s.taskDAO.UpdateStatus(taskID, "running", 0, "", "", &now, nil)
}

func (s *TaskService) CompleteSuccess(taskID int, result interface{}) error {
	resultJSON := ""
	if result != nil {
		if b, err := json.Marshal(result); err == nil {
			resultJSON = string(b)
		}
	}
	now := time.Now()
	return s.taskDAO.UpdateStatus(taskID, "success", 100, resultJSON, "", nil, &now)
}

func (s *TaskService) CompleteFailure(taskID int, err error) error {
	now := time.Now()
	message := ""
	if err != nil {
		message = err.Error()
	}
	return s.taskDAO.UpdateStatus(taskID, "failed", 100, "", message, nil, &now)
}

func (s *TaskService) UpdateProgress(taskID int, progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	_ = s.taskDAO.UpdateProgress(taskID, progress)
}

func (s *TaskService) Cancel(taskID int, finishedAt *time.Time) error {
	return s.taskDAO.Cancel(taskID, finishedAt)
}

func (s *TaskService) GetByID(taskID int, userID int) (*models.Task, error) {
	task, err := s.taskDAO.GetByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, errors.New("无权访问该任务")
	}
	return task, nil
}

func (s *TaskService) ListByUser(userID int, page, pageSize int) (*models.TaskListResponse, error) {
	list, total, err := s.taskDAO.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &models.TaskListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *TaskService) ListByUserAndType(userID int, taskType string, page, pageSize int) (*models.TaskListResponse, error) {
	list, total, err := s.taskDAO.ListByUserAndType(userID, taskType, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &models.TaskListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
