package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func MakeTestTaskRequest(r *mux.Router, method, x, token string, b *bytes.Buffer) *http.Response {
	req := httptest.NewRequest(method, "/api-01/task"+x, b)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result()
}

func MakeTestQueueRequest(r *mux.Router, method, x, token string, b *bytes.Buffer) *http.Response {
	req := httptest.NewRequest(method, "/api-01/queue"+x, b)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result()
}

func MakeTestUploadRequest(r *mux.Router, method, x, ct string, b *bytes.Buffer) *http.Response {
	req := httptest.NewRequest(method, "/api-01/upload"+x, b)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result()
}

func Test_Upload(t *testing.T) {
	fmt.Println("Test_Upload")
	logInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	Conf.MaxFiles = 2
	Conf.MinFiles = 2
	Conf.MaxFileSize = 1024 * 100
	Conf.DataDir = "tmp"
	db, err := BoltNewStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")
	token := NewId(16)
	r := setRouting(token, db)
	b := new(bytes.Buffer)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, err := writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 := strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part2, err := writer.CreateFormFile("file", "file2.bin")
	if err != nil {
		t.Errorf("Can't read file2: %s", err)
	}
	file2 := strings.NewReader(NewId(64))
	_, err = io.Copy(part2, file2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp := MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 201 {
		t.Errorf("Status expected 201 but was: %d", resp.StatusCode)
	}
	task := &TTaskAnswer{}
	err = task.fromJReader(resp.Body)
	if err != nil {
		t.Errorf("JSON parser error: %s", err)
	}
	// test task status
	resp = MakeTestTaskRequest(r, "GET", "/"+task.TaskId, "", b)
	if resp.StatusCode != 202 {
		t.Errorf("Status expected 202 but was: %d", resp.StatusCode)
	}
	// test get queue
	resp = MakeTestQueueRequest(r, "GET", "", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	// test ok
	resp = MakeTestTaskRequest(r, "PATCH", "/"+task.TaskId+"/ok", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	// test task status
	resp = MakeTestTaskRequest(r, "GET", "/"+task.TaskId, "", b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	// test empty queue
	resp = MakeTestQueueRequest(r, "GET", "", token, b)
	if resp.StatusCode != 204 {
		t.Errorf("Status expected 204 but was: %d", resp.StatusCode)
	}
	// test retry verify
	resp = MakeTestTaskRequest(r, "PATCH", "/"+task.TaskId+"/fail", token, b)
	if resp.StatusCode != 409 {
		t.Errorf("Status expected 409 but was: %d", resp.StatusCode)
	}
	// cleanup
	_ = os.RemoveAll(Conf.DataDir + "/" + "_")
	_ = os.RemoveAll(Conf.DataDir + "/" + string(task.TaskId[0]))
}

func Test_Upload_Too_Big(t *testing.T) {
	fmt.Println("Test_Upload_Too_Big")
	logInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	Conf.MaxFiles = 2
	Conf.MinFiles = 2
	Conf.MaxFileSize = 65
	Conf.DataDir = "tmp"
	db, err := BoltNewStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")
	r := setRouting("", db)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part2, err := writer.CreateFormFile("file", "file2.bin")
	if err != nil {
		t.Errorf("Can't read file2: %s", err)
	}
	file2 := strings.NewReader(NewId(64))
	_, err = io.Copy(part2, file2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part1, err := writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 := strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp := MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 400 {
		t.Errorf("Status expected 400 but was: %d", resp.StatusCode)
	}
	_ = os.RemoveAll(Conf.DataDir + "/" + "_")
}

func Test_Upload_Too_Many_Files(t *testing.T) {
	fmt.Println("Test_Upload_Too_Many_Files")
	logInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	Conf.MaxFiles = 2
	Conf.MinFiles = 2
	Conf.MaxFileSize = 1024 * 100
	Conf.DataDir = "tmp"
	db, err := BoltNewStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")
	r := setRouting("", db)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, err := writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 := strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part2, err := writer.CreateFormFile("file", "file2.bin")
	if err != nil {
		t.Errorf("Can't read file2: %s", err)
	}
	file2 := strings.NewReader(NewId(64))
	_, err = io.Copy(part2, file2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part3, err := writer.CreateFormFile("file", "file3.bin")
	if err != nil {
		t.Errorf("Can't read file3: %s", err)
	}
	file3 := strings.NewReader(NewId(32))
	_, err = io.Copy(part3, file3)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp := MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 400 {
		t.Errorf("Status expected 400 but was: %d", resp.StatusCode)
	}
	_ = os.RemoveAll(Conf.DataDir + "/" + "_")
}

func Test_Upload_Too_Few_Files(t *testing.T) {
	fmt.Println("Test_Upload_Too_Few_Files")
	logInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	Conf.MaxFiles = 2
	Conf.MinFiles = 2
	Conf.MaxFileSize = 1024 * 100
	Conf.DataDir = "tmp"
	db, err := BoltNewStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")
	r := setRouting("", db)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, err := writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 := strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp := MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 400 {
		t.Errorf("Status expected 400 but was: %d", resp.StatusCode)
	}
	_ = os.RemoveAll(Conf.DataDir + "/" + "_")
}

func Test_Purge(t *testing.T) {
	fmt.Println("Test_Purge")
	logInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	Conf.MaxFiles = 2
	Conf.MinFiles = 2
	Conf.MaxFileSize = 1024 * 100
	Conf.DataDir = "tmp"
	db, err := BoltNewStorage("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")
	token := NewId(16)
	r := setRouting(token, db)
	b := new(bytes.Buffer)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part1, err := writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 := strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part2, err := writer.CreateFormFile("file", "file2.bin")
	if err != nil {
		t.Errorf("Can't read file2: %s", err)
	}
	file2 := strings.NewReader(NewId(64))
	_, err = io.Copy(part2, file2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp := MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 201 {
		t.Errorf("Status expected 201 but was: %d", resp.StatusCode)
	}
	task := &TTaskAnswer{}
	err = task.fromJReader(resp.Body)
	if err != nil {
		t.Errorf("JSON parser error: %s", err)
	}
	// test task status
	resp = MakeTestTaskRequest(r, "GET", "/"+task.TaskId, "", b)
	if resp.StatusCode != 202 {
		t.Errorf("Status expected 202 but was: %d", resp.StatusCode)
	}
	// test get queue
	resp = MakeTestQueueRequest(r, "GET", "", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	// task 2
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part1, err = writer.CreateFormFile("file", "file1.bin")
	if err != nil {
		t.Errorf("Can't read file1: %s", err)
	}
	file1 = strings.NewReader(NewId(128))
	_, err = io.Copy(part1, file1)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	part2, err = writer.CreateFormFile("file", "file2.bin")
	if err != nil {
		t.Errorf("Can't read file2: %s", err)
	}
	file2 = strings.NewReader(NewId(64))
	_, err = io.Copy(part2, file2)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	err = writer.Close()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	resp = MakeTestUploadRequest(r, "POST", "", writer.FormDataContentType(), body)
	if resp.StatusCode != 201 {
		t.Errorf("Status expected 201 but was: %d", resp.StatusCode)
	}
	// task 2
	time.Sleep(5 * time.Second)
	task2 := &TTaskAnswer{}
	err = task2.fromJReader(resp.Body)
	if err != nil {
		t.Errorf("JSON parser error: %s", err)
	}
	// test task status
	resp = MakeTestTaskRequest(r, "GET", "/"+task2.TaskId, "", b)
	if resp.StatusCode != 202 {
		t.Errorf("Status expected 202 but was: %d", resp.StatusCode)
	}
	// test get queue
	resp = MakeTestQueueRequest(r, "GET", "", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	// test  verify
	resp = MakeTestTaskRequest(r, "PATCH", "/"+task.TaskId+"/ok", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	dberr := db.TaskPurge("verified", 2)
	if dberr != nil {
		t.Errorf("Unexpected error: %s", dberr.msg)
	}
	resp = MakeTestQueueRequest(r, "GET", "", token, b)
	if resp.StatusCode != 200 {
		t.Errorf("Status expected 200 but was: %d", resp.StatusCode)
	}
	resp = MakeTestTaskRequest(r, "GET", "/"+task.TaskId, "", b)
	if resp.StatusCode != 400 {
		t.Errorf("Status expected 400 but was: %d", resp.StatusCode)
	}
	// cleanup
	_ = os.RemoveAll(Conf.DataDir + "/" + "_")
	_ = os.RemoveAll(Conf.DataDir + "/" + string(task.TaskId[0]))
	_ = os.RemoveAll(Conf.DataDir + "/" + string(task2.TaskId[0]))
}
