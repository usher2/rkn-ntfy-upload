package main

import (
	"encoding/json"
	"net/http"
)

const (
	E_INVALID_REQUEST string = "invalid_request"
        E_FILE_TOO_BIG           = "file_too_big"
	E_TASK_NOT_FOUND         = "invalid_task"
	E_ACCESS_DENIED          = "invalid_token"
	E_CONFLICT               = "status_conflict"
	E_QUEUE_EMPTY            = "empty_queue"
	E_SERVER_ERROR           = "server_error"
	E_NOT_IMPLEMENTED        = "not_implemented"
)

type TJSONError struct {
	Msg string `json:"error"`
}

func sendJSONErrorMessage(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	c := &TJSONError{msg}
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	e.Encode(c)
}
