package main

import (
	"log"
	"os"
)

var errorLogFile *os.File
var appLogFile *os.File

func init() {

	appLogFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	errorLogFile, err = os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
}

func LogError(err error) {
	// Set the output of the log package to the file
	log.SetOutput(errorLogFile)
	logError := log.New(errorLogFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError.Println(err)
}

func LogApp(msg string) {
	// Set the output of the log package to the file
	log.SetOutput(appLogFile)
	logApp := log.New(appLogFile, "APP: ", log.Ldate|log.Ltime|log.Lshortfile)
	logApp.Println(msg)
}
