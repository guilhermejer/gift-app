package logger

import (
	"encoding/json"
	"log"
	"time"
)

type HTTPRequestEntry struct {
	Level      string `json:"level"`
	Date       string `json:"date"`
	Time       string `json:"time"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Query      string `json:"query,omitempty"`
	RemoteAddr string `json:"remote_addr,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Payload    string `json:"payload,omitempty"`
}

type HTTPResponseEntry struct {
	Level         string `json:"level"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	RequestID     string `json:"request_id"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	Status        int    `json:"status"`
	LatencyMs     int64  `json:"latency_ms"`
	ResponseBytes int    `json:"response_bytes"`
	Payload       string `json:"payload,omitempty"`
}

func LogRequest(requestID, method, path, query, remoteAddr, userAgent, payload string) {
	now := time.Now()
	entry := HTTPRequestEntry{
		Level:      LevelInfo,
		Date:       now.Format("2006-01-02"),
		Time:       now.Format("15:04:05"),
		Message:    "request started",
		RequestID:  requestID,
		Method:     method,
		Path:       path,
		Query:      query,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
		Payload:    payload,
	}
	logEntry(entry)
}

func LogResponse(requestID, method, path string, status int, latency time.Duration, responseBytes int, payload string) {
	now := time.Now()
	entry := HTTPResponseEntry{
		Level:         LevelInfo,
		Date:          now.Format("2006-01-02"),
		Time:          now.Format("15:04:05"),
		Message:       "request completed",
		RequestID:     requestID,
		Method:        method,
		Path:          path,
		Status:        status,
		LatencyMs:     latency.Milliseconds(),
		ResponseBytes: responseBytes,
		Payload:       payload,
	}
	logEntry(entry)
}

func logEntry(entry any) {
	payload, err := json.Marshal(entry)
	if err != nil {
		fallback := Entry{
			Level:   LevelError,
			Date:    time.Now().Format("2006-01-02"),
			Time:    time.Now().Format("15:04:05"),
			Message: "failed to marshal log entry",
			Error:   err.Error(),
		}
		fallbackPayload, _ := json.Marshal(fallback)
		log.Println(string(fallbackPayload))
		return
	}

	log.Println(string(payload))
}
