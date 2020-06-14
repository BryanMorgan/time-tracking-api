package task

import (
	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/valid"
)

// Compile Only: ensure interface is implemented
var _ TaskService = &TaskResource{}

type TaskService interface {
	GetTask(taskId int, accountId int) (*Task, *api.Error)
	GetAllTasks(accountId int, active bool) ([]*Task, *api.Error)

	SaveTask(accountId int, name string, common bool, rate float64, billable bool) (*Task, *api.Error)

	UpdateTask(*Task) *api.Error

	ArchiveTask(taskId int, accountId int) *api.Error
	RestoreTask(taskId int, accountId int) *api.Error
	DeleteTask(taskId int, accountId int) *api.Error
}

type TaskResource struct {
	store TaskStore
}

func NewTaskService(store TaskStore) TaskService {
	return &TaskResource{store: store}
}

func (c *TaskResource) GetTask(taskId int, accountId int) (*Task, *api.Error) {
	task, err := c.store.GetTask(taskId, accountId)
	if err != nil {
		return nil, api.NewError(err, "Could not get task", api.SystemError)
	}

	return task, nil
}

func (c *TaskResource) GetAllTasks(accountId int, active bool) ([]*Task, *api.Error) {
	task, err := c.store.GetAllTasks(accountId, active)
	if err != nil {
		return nil, api.NewError(err, "Could not get all tasks", api.SystemError)
	}

	return task, nil
}

func (c *TaskResource) SaveTask(accountId int, name string, common bool, rate float64, billable bool) (*Task, *api.Error) {
	newTask := Task{
		AccountId:       accountId,
		Name:            name,
		DefaultRate:     valid.ToNullFloat64(rate),
		DefaultBillable: billable,
		Common:          common,
	}
	taskId, err := c.store.SaveTask(&newTask)
	if err != nil {
		return nil, api.NewError(err, "Failed to save task", api.SystemError)
	}

	newTask.TaskId = taskId
	return &newTask, nil
}

func (c *TaskResource) UpdateTask(updateTask *Task) *api.Error {
	err := c.store.UpdateTask(updateTask)
	if err != nil {
		return api.NewError(err, "Failed to update task", api.SystemError)
	}

	return nil
}

func (c *TaskResource) ArchiveTask(taskId int, accountId int) *api.Error {
	err := c.store.ArchiveTask(taskId, accountId)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "Task not found", api.InvalidTask)
	} else if err != nil {
		return api.NewError(err, "Failed to archive task", api.SystemError)
	}

	return nil
}

func (c *TaskResource) RestoreTask(taskId int, accountId int) *api.Error {
	err := c.store.RestoreTask(taskId, accountId)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "Task not found", api.InvalidTask)
	} else if err != nil {
		return api.NewError(err, "Failed to restore archived task", api.SystemError)
	}

	return nil
}

func (c *TaskResource) DeleteTask(taskId int, accountId int) *api.Error {
	err := c.store.DeleteTask(taskId, accountId)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "Task not found", api.InvalidTask)
	} else if err != nil {
		return api.NewError(err, "Failed to delete task", api.SystemError)
	}

	return nil
}

// Find tasks that are in the source list, but not in the other list
func GetTaskListDifference(source []Task, other []Task) []Task {
	var results []Task

	for _, sourceTask := range source {
		match := false
		for _, otherTask := range other {
			if sourceTask.TaskId == otherTask.TaskId {
				match = true
				break
			}
		}

		if !match {
			results = append(results, Task{TaskId: sourceTask.TaskId})
		}
	}
	return results
}
