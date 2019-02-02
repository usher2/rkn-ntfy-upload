package main

import (
	"io"
	"log"
)

var (
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func logInit(debugHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Debug = log.New(debugHandle, "[=]: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	Info = log.New(infoHandle, "[+]: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	Warning = log.New(warningHandle, "[!]: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	Error = log.New(errorHandle, "[-]: ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
}
