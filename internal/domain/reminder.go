package domain

import "time"

type Reminder struct {
	ReminderID string
	UserID     string
	FriendID   string
	Type       string
	TriggerAt  time.Time
	Message    string
}
