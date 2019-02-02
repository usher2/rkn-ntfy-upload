package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"time"
        "regexp"
)

const (
	TASK_ID_LEN = 32
	TEMPDIR_LEN = 32
)

type TTask struct {
	Id       string `json:"id"`               // a unique identifier
	Status   string `json:"status,omitempty"` // upload status "", "received", "verified", "invalid"
	IssuedAt int64  `json:"iat"`              // issued time
}

type TTaskAnswer struct {
	TaskId string `json:"task"` // a unique identifier
}

type TTaskStatus struct {
	Status string `json:"status"`
}

// Fill TTask object from byte array
func (c *TTask) fromJBytes(b []byte) error {
	return json.Unmarshal(b, c)
}

// Make byte array from TTask object
func (c *TTask) toJBytes() ([]byte, error) {
	return json.Marshal(c)
}

// Write TTask object to io.Writer as JSON
func (c *TTask) toJWriter(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	return e.Encode(c)
}

// Write TTaskStatus object to io.Writer as JSON
func (c *TTaskStatus) toJWriter(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	return e.Encode(c)
}

// Write TTaskAnswer object to io.Writer as JSON
func (c *TTaskAnswer) toJWriter(w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	return e.Encode(c)
}

// Fill TTaskAnswer object from io.Reader as JSON (only fow test
func (c *TTaskAnswer) fromJReader(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

// complete task
func taskCompleteHandler(w http.ResponseWriter, r *http.Request, db IStorage, status string) {
	// implies, that the method and content type checks was completed at the routing stage
	var task TTask
	var err error
	vars := mux.Vars(r)
	task_id := vars["task"]
	// get data from the database
	oldTaskPayload, dberr := db.TaskGet(task_id)
	if dberr != nil {
		if dberr.code == E_STORAGE_TASK_NOT_FOUND {
			sendJSONErrorMessage(w, E_TASK_NOT_FOUND, http.StatusBadRequest)
			Warning.Printf("[%s]: Task not found: %s\n", r.RemoteAddr, task_id)
		} else if dberr.code == E_STORAGE_DATABASE_ERROR {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Database error: %s\n", r.RemoteAddr, dberr)
		}
		return
	}
	err = task.fromJBytes(oldTaskPayload)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: JSON syntax error in database payload: %s\n", r.RemoteAddr, err)
		return
	}
	if task.Status != "received" {
		sendJSONErrorMessage(w, E_CONFLICT, http.StatusConflict)
		Warning.Printf("[%s]: Task conflict: %s\n", r.RemoteAddr, task_id)
	}
	if status == "ok" {
		task.Status = "verified"
	} else if status == "fail" {
		task.Status = "failed"
	}
	newTaskPayload, err := task.toJBytes()
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: JSON syntax error payload: %s\n", r.RemoteAddr, err)
		return
	}
	// update the Task in the database
	dberr = db.TaskComplete(task_id, oldTaskPayload, newTaskPayload)
	if dberr != nil {
		if dberr.code == E_STORAGE_TASK_CONFLICT {
			sendJSONErrorMessage(w, E_CONFLICT, http.StatusConflict)
			Warning.Printf("[%s]: Task conflict: %s\n", r.RemoteAddr, task_id)
		} else if dberr.code == E_STORAGE_TASK_NOT_FOUND {
			sendJSONErrorMessage(w, E_TASK_NOT_FOUND, http.StatusBadRequest)
			Warning.Printf("[%s]: Task not found: %s\n", r.RemoteAddr, task_id)
		} else if dberr.code == E_STORAGE_DATABASE_ERROR {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Database error: %s\n", r.RemoteAddr, dberr)
		}
		return
	}
	// start a normal output
	HelperSetStandartHeaders(w)
	w.WriteHeader(http.StatusOK)
	// write a Task info
	err = task.toJWriter(w)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
		return
	}
	// write a success message to the log
	Info.Printf("[%s]: Task completed: %s (%s)\n", r.RemoteAddr, task_id, status)
}

// taskInfoHandler outputs the Task info
func taskInfoHandler(w http.ResponseWriter, r *http.Request, db IStorage) {
	// implies, that the method and content type checks was completed at the routing stage
	var task TTask
	var err error
	vars := mux.Vars(r)
	task_id := vars["task"]
	// get data from the database
	payload, dberr := db.TaskGet(task_id)
	if dberr != nil {
		if dberr.code == E_STORAGE_TASK_NOT_FOUND {
			sendJSONErrorMessage(w, E_TASK_NOT_FOUND, http.StatusBadRequest)
			Warning.Printf("[%s]: Task not found: %s\n", r.RemoteAddr, task_id)
		} else if dberr.code == E_STORAGE_DATABASE_ERROR {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Database error: %s\n", r.RemoteAddr, dberr)
		}
		return
	}
	// validate a data format
	err = task.fromJBytes(payload)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: JSON syntax error in database payload: %s\n", r.RemoteAddr, err)
		return
	}
	status := &TTaskStatus{}
	if task.Status == "verified" {
		status.Status = "ok"
	} else if task.Status == "failed" {
		status.Status = "failed"
	} else if task.Status == "received" {
		status.Status = "wait"
	} else {
		sendJSONErrorMessage(w, E_TASK_NOT_FOUND, http.StatusBadRequest)
		Warning.Printf("[%s]: Task found, but not uploads: %s\n", r.RemoteAddr, task_id)
		return
	}
	// start a normal output
	HelperSetStandartHeaders(w)
	if status.Status == "wait" {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	// write a Task status
	err = status.toJWriter(w)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
		return
	}
	// write a success message to the debug log
	Debug.Printf("[%s]: Task %s info printed\n", r.RemoteAddr, task_id)
}

// queueFirstHandler outputs the first Task info
func queueFirstHandler(w http.ResponseWriter, r *http.Request, db IStorage) {
	// implies, that the method and content type checks was completed at the routing stage
	var task TTask
	var err error
	// get data from the database
	payload, dberr := db.QueueGet()
	if dberr != nil {
		if dberr.code == E_STORAGE_QUEUE_IS_EMPTY {
			sendJSONErrorMessage(w, E_QUEUE_EMPTY, http.StatusNoContent)
			Debug.Printf("[%s]: Queue is empty\n", r.RemoteAddr)
		} else if dberr.code == E_STORAGE_DATABASE_ERROR {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Database error: %s\n", r.RemoteAddr, dberr)
		}
		return
	}
	// validate a data format
	err = task.fromJBytes(payload)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: JSON syntax error in database payload: %s\n", r.RemoteAddr, err)
		return
	}
	// start a normal output
	HelperSetStandartHeaders(w)
	w.WriteHeader(http.StatusOK)
	// write a Task info
	err = task.toJWriter(w)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
		return
	}
	// write a success message to the debug log
	Debug.Printf("[%s]: Task %s info printed for queue\n", r.RemoteAddr, task.Id)
}

// upload files
func uploadHandler(w http.ResponseWriter, r *http.Request, db IStorage) {
	var task TTask
	task.Id = NewId(TASK_ID_LEN)
	task.IssuedAt = time.Now().Unix()
	task.Status = "received"
	reader, err := r.MultipartReader()
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
		return
	}
	// Create unique upload dir with underline and than rename it
	// Defer cleanup those
	dataDir := Conf.DataDir + "/" + "_" + "/" + task.Id
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Temp path exists: %s\n", r.RemoteAddr, task.Id)
		return
	}
	dirCleanup := func() {
		_ = os.RemoveAll(dataDir)
	}
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Can't create temporary directory: %s\n", r.RemoteAddr, err)
		return
	}
	defer dirCleanup()
        re := regexp.MustCompile(`\s*\(\d+\)\s*\.`)
	fcounter := int(0) // uploaded file counter
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if part.FileName() != "" {
			if fcounter >= Conf.MaxFiles {
				sendJSONErrorMessage(w, E_INVALID_REQUEST, http.StatusBadRequest)
				Error.Printf("[%s]: Too many files\n", r.RemoteAddr)
				return
			}
                        _filename := re.ReplaceAllString(part.FileName(),".")
			dst, err := os.Create(dataDir + "/" + _filename)
			if err != nil {
				sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
				Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
				return
			}
			defer dst.Close()
			n, err := io.Copy(dst, io.LimitReader(part, Conf.MaxFileSize))
			if err == nil && n == Conf.MaxFileSize {
				sendJSONErrorMessage(w, E_INVALID_REQUEST, http.StatusBadRequest)
				Error.Printf("[%s]: File too large: %s\n", r.RemoteAddr, _filename)
				return
			} else if err != nil && err != io.EOF {
				sendJSONErrorMessage(w, E_INVALID_REQUEST, http.StatusBadRequest)
				Error.Printf("[%s]: Can't read file %s: %s\n", r.RemoteAddr, _filename, err)
				return
			}
			dst.Close()
			fcounter++
		}
	}
	// check min settings
	if fcounter < Conf.MinFiles {
		sendJSONErrorMessage(w, E_INVALID_REQUEST, http.StatusBadRequest)
		Error.Printf("[%s]: Too few files\n", r.RemoteAddr)
		return
	}
	taskPayload, err := task.toJBytes()
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: JSON syntax error payload: %s\n", r.RemoteAddr, err)
		return
	}
	dberr := db.TaskQueue(task.Id, taskPayload)
	if dberr != nil {
		if dberr.code == E_STORAGE_TASK_EXISTS {
			sendJSONErrorMessage(w, E_CONFLICT, http.StatusConflict)
			Error.Printf("[%s]: Task status conflict\n", r.RemoteAddr)
			return
		} else {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Database error: %s\n", r.RemoteAddr, dberr)
			return
		}
	}
	// Rename dir if can
	newDir := Conf.DataDir + "/" + string(task.Id[0]) + "/" + string(task.Id[1])
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		err = os.MkdirAll(newDir, 0755)
		if err != nil {
			sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
			Error.Printf("[%s]: Can't create new directory: %s\n", r.RemoteAddr, err)
			return
		}
	}
	err = os.Rename(dataDir, newDir+"/"+task.Id)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Can't rename directory: %s\n", r.RemoteAddr, err)
		return
	}
	answer := &TTaskAnswer{TaskId: task.Id}
	// start a normal output
	HelperSetStandartHeaders(w)
	w.WriteHeader(http.StatusCreated)
	// write a Queue info
	err = answer.toJWriter(w)
	if err != nil {
		sendJSONErrorMessage(w, E_SERVER_ERROR, http.StatusInternalServerError)
		Error.Printf("[%s]: Unexpected error: %s\n", r.RemoteAddr, err)
		return
	}
	// write a success message to the log
	Info.Printf("[%s]: Files uploaded\n", r.RemoteAddr)
}

// handler for OPTIONS request
func optionsHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Header.Get("Access-Control-Request-Method")
	headers := r.Header.Get("Access-Control-Request-Headers")
	if method != "POST" && method != "GET" && method != "DELETE" && method != "PUT" && method != "PATCH" {
		method = "POST"
	}
	// start output
	w.Header().Set("Access-Control-Allow-Method", method)
	w.Header().Set("Access-Control-Allow-Headers", headers)
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	// write a success message to the log
	Info.Printf("[%s]: OPTIONS\n", r.RemoteAddr)
}

func HelperSetStandartHeaders(w http.ResponseWriter) {
	w.Header().Add(
		"Cache-Control",
		"no-cache, no-store, max-age=0, must-revalidate",
	)
	w.Header().Add("Pragma", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	return
}
