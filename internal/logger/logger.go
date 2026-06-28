package logger

import (
	"encoding/json"
	"errors"
	"log"
	"time"
)

const (
	LevelError = "error"
)

type Entry struct {
	Level   string `json:"level"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func Log(level, message string, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}

	now := time.Now()
	entry := Entry{
		Level:   level,
		Date:    now.Format("2006-01-02"),
		Time:    now.Format("15:04:05"),
		Message: message,
		Error:   err.Error(),
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		fallback := Entry{
			Level:   LevelError,
			Date:    now.Format("2006-01-02"),
			Time:    now.Format("15:04:05"),
			Message: "failed to marshal log entry",
			Error:   err.Error(),
		}
		fallbackPayload, _ := json.Marshal(fallback)
		log.Println(string(fallbackPayload))
		return
	}

	log.Println(string(payload))
}
