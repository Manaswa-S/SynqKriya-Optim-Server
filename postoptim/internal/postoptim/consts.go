package postoptim

import (
	"fmt"
	"log"
	"os"
)

func newErr(err error) {
	fmt.Println(err)
}

type LogSeverity string

const (
	LogInfo   LogSeverity = "INFO"
	LogWarn   LogSeverity = "WARN"
	LogError  LogSeverity = "ERROR"
	LogMetric LogSeverity = "METRIC"
)

func (s LogSeverity) String() string {
	return string(s)
}

var logFile *os.File

// newLog logs an event at all required places.
// Callers are explicitly required to include special chars like new-line.
func newLog(severity LogSeverity, logStr string) {
	fmt.Print(severity.String() + ": " + logStr) // this intentionally does not put a new line at the end

	if logFile == nil {
		var err error
		regularLogPath := "./logs/postoptim.log"
		err = os.MkdirAll(regularLogPath, 0644)
		if err != nil {
			log.Fatalf("failed to create postoptim.log : %v", err)
		}

		logFile, err = os.OpenFile(regularLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open postoptim.log : %v", err)
		}
	}

	logFile.WriteString(severity.String() + ": " + logStr)
}
