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

// InitLogger inicializa o logger com configurações padrão
func InitLogger(enabled *bool) {
	isEnabled = enabled
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

// LogDebug registra uma mensagem de debug
func LogDebug(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("DEBUG: %s", msg)
	}
}

// LogError registra uma mensagem de erro
func LogError(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("ERROR: %s", msg)
	}
}

// LogInfo registra uma mensagem informativa
func LogInfo(format string, v ...interface{}) {
	if isEnabled != nil && *isEnabled {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("INFO: %s", msg)
	}
}

// LogPerformance registra informações de performance
func LogPerformance(operation string, duration time.Duration) {
	if isEnabled != nil && *isEnabled {
		logger.Printf("PERFORMANCE: %s took %v", operation, duration)
	}
}
