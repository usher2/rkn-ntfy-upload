package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type LocalConfig struct {
	LogLevel        string // "Debug", "Info", "Warning", "Error"
	DataDir         string // = "tmp"
	MaxFiles        int    // = 2
	MinFiles        int    // = 2
	MaxFileSize     int64  // = 1024 * 100
	CompleteTaskTTL int64  // = 3600 * 24
	ListenPort      string // = 14000
	AuthToken       string // = "12313425435345"
	DataBaseFile    string // = "my.db"
}

var Conf LocalConfig

func main() {
	flag.StringVar(&Conf.DataDir, "d", "tmp", "Data directory")
	flag.StringVar(&Conf.LogLevel, "l", "Debug", "Logging Level (Debug, Info, Warning, Error)")
	flag.IntVar(&Conf.MaxFiles, "a", 2, "Maximum number of files")
	flag.IntVar(&Conf.MinFiles, "i", 2, "Minimum number of files")
	flag.Int64Var(&Conf.MaxFileSize, "s", 1024*1024, "Maximum file size")
	flag.Int64Var(&Conf.CompleteTaskTTL, "c", 3600, "Task TTL after verification upload")
	flag.StringVar(&Conf.ListenPort, "p", "14000", "Listen port")
	flag.StringVar(&Conf.AuthToken, "x", "12313425435345", "Auth token")
	flag.StringVar(&Conf.DataBaseFile, "b", "my.db", "Database file")
	flag.Parse()
	if Conf.LogLevel == "Info" {
		logInit(ioutil.Discard, os.Stdout, os.Stderr, os.Stderr)
	} else if Conf.LogLevel == "Warning" {
		logInit(ioutil.Discard, ioutil.Discard, os.Stderr, os.Stderr)
	} else if Conf.LogLevel == "Error" {
		logInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		logInit(os.Stderr, os.Stdout, os.Stderr, os.Stderr)
	}
	db, err := BoltNewStorage(Conf.DataBaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	go func() {
		db.TaskPurge("verified", Conf.CompleteTaskTTL)
		db.TaskPurge("failed", Conf.CompleteTaskTTL)
		time.Sleep(10 * time.Second)
	}()
	r := setRouting(Conf.AuthToken, db)
	http.Handle("/", r)
	listen := ":" + Conf.ListenPort
	http.ListenAndServe(listen, nil)
}
