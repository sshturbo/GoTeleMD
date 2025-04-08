package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logger    *log.Logger
	isEnabled *bool
)

func InitLogger(enabled *bool) {
	isEnabled = enabled
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func LogDebug(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("DEBUG: %s", msg)
	}
}

func LogError(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("ERROR: %s", msg)
	}
}

func LogInfo(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("INFO: %s", msg)
	}
}

func LogPerformance(operation string, duration time.Duration) {
	if isEnabled != nil && *isEnabled {
		logger.Printf("PERFORMANCE: %s took %v", operation, duration)
	}
}
