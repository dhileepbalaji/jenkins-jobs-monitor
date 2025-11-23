package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Logger is a custom logger for structured JSON logging
type Logger struct {
	mu     sync.Mutex
	logger *log.Logger
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Message   string `json:"message"`
}

var globalLogger *Logger

// SetupLogging initializes the global logger to write to both a file and console
func SetupLogging(logFilePath string) {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v. Logging to stdout only.", logFilePath, err)
		} else {
			writers = append(writers, file)
		}
	}

	// Create a multi-writer to write to both file and console
	multiWriter := io.MultiWriter(writers...)

	globalLogger = &Logger{
		logger: log.New(multiWriter, "", 0), // No default flags, we'll format manually
	}
}

// Info logs an informational message
func Info(message string) {
	if globalLogger == nil {
		log.Println("INFO: " + message)
		return
	}
	globalLogger.log("INFO", message)
}

// Error logs an error message
func Error(message string) {
	if globalLogger == nil {
		log.Println("ERROR: " + message)
		return
	}
	globalLogger.log("ERROR", message)
}

// Fatal logs a fatal error message and exits
func Fatal(message string) {
	if globalLogger == nil {
		log.Println("FATAL: " + message)
		os.Exit(1)
	}
	globalLogger.log("FATAL", message)
	os.Exit(1)
}

func (l *Logger) log(level, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, file, line, _ := runtime.Caller(2) // Caller(2) to get the original caller of Info/Error/Fatal
	shortFile := filepath.Base(file)

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		File:      shortFile,
		Line:      line,
		Message:   message,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		l.logger.Printf("Error marshalling log entry: %v, Original Message: %s", err, message)
		return
	}
	l.logger.Println(string(jsonBytes))
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func GetDir(path string) string {
	return filepath.Dir(path)
}

func PrintLine(length int) {
	fmt.Println(strings.Repeat("-", length))
}
