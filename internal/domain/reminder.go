package domain

import "time"

type ReminderType string

const (
	ReminderTypeBirthday ReminderType = "birthday"
	ReminderTypeCustom   ReminderType = "custom"
)

type Reminder struct {
	ReminderID string
	UserID     string
	FriendID   string
	Type       ReminderType
	TriggerAt  time.Time
	Message    string
}
