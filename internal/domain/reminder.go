package domain

import "time"

type ReminderRecurrence string

const (
	ReminderRecurrenceNone    ReminderRecurrence = "none"
	ReminderRecurrenceYearly  ReminderRecurrence = "yearly"
	ReminderRecurrenceMonthly ReminderRecurrence = "monthly"
	ReminderRecurrenceWeekly  ReminderRecurrence = "weekly"
	ReminderRecurrenceDaily   ReminderRecurrence = "daily"
)

func (r ReminderRecurrence) IsValid() bool {
	switch r {
	case ReminderRecurrenceNone, ReminderRecurrenceYearly, ReminderRecurrenceMonthly, ReminderRecurrenceWeekly, ReminderRecurrenceDaily:
		return true
	}
	return false
}

func (r ReminderRecurrence) IsRecurring() bool {
	return r != "" && r != ReminderRecurrenceNone
}

type Reminder struct {
	ReminderID     string             `json:"reminderID" example:"d8c8efdf-c52f-4d6b-8e2e-b83f78de4f77"`
	UserID         string             `json:"userID" example:"5581a365-394f-467d-ae13-0d01e4cf1863"`
	FriendID       string             `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	Type           string             `json:"type" example:"birthday"`
	TriggerAt      time.Time          `json:"triggerAt" format:"date-time" example:"2026-08-15T00:00:00Z"`
	Recurrence     ReminderRecurrence `json:"recurrence" example:"yearly"`
	Message        string             `json:"message" example:"Comprar presente ate uma semana antes"`
	NextOccurrence *time.Time         `json:"nextOccurrence,omitempty" format:"date-time" example:"2027-08-15T00:00:00Z"`
}
