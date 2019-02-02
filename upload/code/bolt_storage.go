package main

import (
	"encoding/binary"
	"fmt"
	bolt "github.com/etcd-io/bbolt"
	"time"
)

type TBoltStorage struct {
	db *bolt.DB
}

func BoltNewStorage(dbfilename string) (*TBoltStorage, error) {
	db, err := bolt.Open(dbfilename, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("QUEUE"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("TQREL"))
		if err != nil {
			return err
		}
		return nil
	})
	return &TBoltStorage{db}, err
}

// put Task in queue
func (s *TBoltStorage) TaskQueue(taskId string, taskPayload []byte) (err *TErrorStorage) {
	_err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TASKS"))
		bq := tx.Bucket([]byte("QUEUE"))
		btq := tx.Bucket([]byte("TQREL"))
		v := b.Get([]byte(taskId))
		if v != nil {
			return &TErrorStorage{"Task not found", E_STORAGE_TASK_EXISTS}
		}
		qid, _ := bq.NextSequence()
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(qid))
		err := bq.Put(buf, taskPayload)
		if err != nil {
			return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
		}
		err = b.Put([]byte(taskId), taskPayload)
		if err != nil {
			return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
		}
		err = btq.Put([]byte(taskId), buf)
		if err != nil {
			return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
		}
		return nil
	})
	err, _ = _err.(*TErrorStorage)
	return
}

// complete Task
func (s *TBoltStorage) TaskComplete(taskId string, oldTaskPayload, newTaskPayload []byte) (err *TErrorStorage) {
	_err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TASKS"))
		bq := tx.Bucket([]byte("QUEUE"))
		btq := tx.Bucket([]byte("TQREL"))
		v := b.Get([]byte(taskId))
		if v == nil {
			return &TErrorStorage{"Task not found", E_STORAGE_TASK_NOT_FOUND}
		}
		if string(v) != string(oldTaskPayload) {
			return &TErrorStorage{"Task conflict", E_STORAGE_TASK_CONFLICT}
		}
		buf := btq.Get([]byte(taskId))
		if buf != nil {
			err := bq.Delete(buf)
			if err != nil {
				return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
			}
			err = btq.Delete([]byte(taskId))
			if err != nil {
				return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
			}
		}
		err := b.Delete([]byte(taskId))
		if err != nil {
			return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
		}
		err = b.Put([]byte(taskId), newTaskPayload)
		if err != nil {
			return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
		}
		return nil
	})
	err, _ = _err.(*TErrorStorage)
	return
}

func (s *TBoltStorage) TaskGet(taskId string) (taskPayload []byte, err *TErrorStorage) {
	_err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TASKS"))
		taskPayload = b.Get([]byte(taskId))
		if taskPayload == nil {
			return &TErrorStorage{"Task not found", E_STORAGE_TASK_NOT_FOUND}
		}
		return nil
	})
	err, _ = _err.(*TErrorStorage)
	return
}

func (s *TBoltStorage) QueueGet() (taskPayload []byte, err *TErrorStorage) {
	_err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("QUEUE"))
		c := b.Cursor()
		_, taskPayload = c.First()
		if taskPayload == nil {
			return &TErrorStorage{"Queue is empty", E_STORAGE_QUEUE_IS_EMPTY}
		}
		return nil
	})
	err, _ = _err.(*TErrorStorage)
	return
}

func (s *TBoltStorage) TaskPurge(status string, duration int64) (err *TErrorStorage) {
	task := &TTask{}
	t := time.Now().Unix()
	_err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TASKS"))
		bq := tx.Bucket([]byte("QUEUE"))
		btq := tx.Bucket([]byte("TQREL"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := task.fromJBytes(v)
			if err != nil {
				return &TErrorStorage{fmt.Sprintf("Invalid task format: %s", err), E_STORAGE_DATABASE_ERROR}
			}
			if task.Status == status {
				if task.IssuedAt+duration < t {
					buf := btq.Get([]byte(task.Id))
					if buf != nil {
						err = bq.Delete(buf)
						if err != nil {
							return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
						}
						err = btq.Delete([]byte(task.Id))
						if err != nil {
							return &TErrorStorage{fmt.Sprintf("Database error: %s", err), E_STORAGE_DATABASE_ERROR}
						}
					}
					b.Delete(k)
				}
			}
		}
		return nil
	})
	err, _ = _err.(*TErrorStorage)
	return
}

func (s *TBoltStorage) Close() {
	s.db.Close()
}
