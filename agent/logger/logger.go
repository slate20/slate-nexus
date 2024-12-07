package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var LogFile *os.File

func SetupLogger() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get the directory of the executable: %w", err)
	}
	dir := filepath.Dir(exe)

	// Define the path to the log file
	logFilePath := filepath.Join(dir, "agent.log")
	LogFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}

	log.SetOutput(LogFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}

func CustomLog(level, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)

	// Write the log message to the log file
	log.Print(logMessage)
}

func LogInfo(format string, v ...interface{}) {
	CustomLog("INFO", format, v...)
}

func LogError(format string, v ...interface{}) {
	CustomLog("ERROR", format, v...)
}

func LogWarn(format string, v ...interface{}) {
	CustomLog("WARN", format, v...)
}
