package logger

import (
	"fmt"
	"os"
	"time"
)

func LogError(message string) {
	f, err := os.OpenFile("monitor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Printf("Can't open log file: $v\n", err)
		return
	}

	defer f.Close()

	logLine := fmt.Sprintf("%s [ERROR] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)

	if _, err := f.WriteString(logLine); err != nil {
		fmt.Printf("Can't write to file %v\n", err)
	}
}
