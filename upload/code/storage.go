package main

const (
	E_STORAGE_DATABASE_ERROR = iota
	E_STORAGE_TASK_EXISTS
	E_STORAGE_TASK_NOT_FOUND
	E_STORAGE_TASK_CONFLICT
	E_STORAGE_QUEUE_IS_EMPTY
)

type (
	TErrorStorage struct {
		msg  string
		code int
	}

	IStorage interface {
		TaskGet(taskId string) (taskPayload []byte, err *TErrorStorage)
		TaskQueue(taskId string, taskPayload []byte) (err *TErrorStorage)
		TaskComplete(taskId string, oldTaskPayload, newTaskPayload []byte) (err *TErrorStorage)
		TaskPurge(status string, duration int64) (err *TErrorStorage)
		QueueGet() (taskPayload []byte, err *TErrorStorage)
		Close()
	}
)

func (e *TErrorStorage) Error() string { return e.msg }
