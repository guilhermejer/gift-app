package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gift-app/api/internal/domain"
	"github.com/gift-app/api/internal/port"
)

type SuggestionJob struct {
	reminderRepo port.ReminderRepository
	userRepo     port.UserRepository
	giftRepo     port.GiftRepository
	jobLogRepo   port.SuggestionJobLogRepository
	llmClient    interface {
		SuggestionCreate(ctx context.Context, requestID, friendID string, payload interface{}) (map[string]interface{}, error)
	}
	logger   *slog.Logger
	interval time.Duration
}

func NewSuggestionJob(
	reminderRepo port.ReminderRepository,
	userRepo port.UserRepository,
	giftRepo port.GiftRepository,
	jobLogRepo port.SuggestionJobLogRepository,
	llmClient interface {
		SuggestionCreate(ctx context.Context, requestID, friendID string, payload interface{}) (map[string]interface{}, error)
	},
	interval time.Duration,
) *SuggestionJob {
	return &SuggestionJob{
		reminderRepo: reminderRepo,
		userRepo:     userRepo,
		giftRepo:     giftRepo,
		jobLogRepo:   jobLogRepo,
		llmClient:    llmClient,
		logger:       slog.Default().With("component", "suggestion_job"),
		interval:     interval,
	}
}

func (j *SuggestionJob) Run(ctx context.Context) {
	j.logger.Info("suggestion job started", "interval", j.interval)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("suggestion job stopped")
			return
		case <-ticker.C:
			j.execute(ctx)
		}
	}
}

func (j *SuggestionJob) execute(ctx context.Context) {
	now := time.Now().UTC()
	logger := j.logger.With("tick", now.Format(time.RFC3339))
	logger.Info("suggestion job tick start")

	users, err := j.userRepo.ListAll(ctx)
	if err != nil {
		logger.Error("failed to list users", "error", err)
		return
	}

	for _, user := range users {
		lookahead := user.SuggestionLookaheadDays
		if !domain.IsValidLookaheadDays(lookahead) {
			lookahead = domain.DefaultLookaheadDays
		}
		windowEnd := now.AddDate(0, 0, lookahead)

		userLogger := logger.With("user_id", user.UserID, "lookahead", lookahead)

		// Non-recurring reminders with trigger_at in the window
		pending, err := j.reminderRepo.ListPending(ctx, now, windowEnd)
		if err != nil {
			userLogger.Error("failed to list pending reminders", "error", err)
			continue
		}

		var allReminders []*domain.Reminder
		allReminders = append(allReminders, pending...)

		// Recurring reminders expanded via OccurrencesBetween
		recurring, err := j.reminderRepo.ListRecurring(ctx)
		if err != nil {
			userLogger.Error("failed to list recurring reminders", "error", err)
			continue
		}
		for _, rem := range recurring {
			occurrences := domain.OccurrencesBetween(rem.Recurrence, rem.TriggerAt, now, windowEnd)
			for _, occ := range occurrences {
				r := *rem
				r.TriggerAt = occ
				allReminders = append(allReminders, &r)
			}
		}

		// Filter to this user's reminders
		for _, rem := range allReminders {
			if rem.UserID != user.UserID {
				continue
			}
			occDate := rem.TriggerAt
			j.processReminder(ctx, userLogger, rem, occDate, now, lookahead)
		}
	}

	logger.Info("suggestion job tick done")
}

func (j *SuggestionJob) processReminder(
	ctx context.Context,
	logger *slog.Logger,
	rem *domain.Reminder,
	occDate time.Time,
	now time.Time,
	lookahead int,
) {
	occDateTrunc := occDate.Truncate(24 * time.Hour)
	remLogger := logger.With("reminder_id", rem.ReminderID, "friend_id", rem.FriendID, "occ_date", occDateTrunc.Format("2006-01-02"))

	exists, err := j.jobLogRepo.Exists(ctx, rem.ReminderID, occDateTrunc)
	if err != nil {
		remLogger.Error("failed to check job log", "error", err)
		return
	}
	if exists {
		remLogger.Debug("already processed, skipping")
		return
	}

	// Check if recent gifts already exist for this reminder
	since := now.AddDate(0, 0, -lookahead)
	recentGifts, err := j.giftRepo.ListRecentByReminderID(ctx, rem.ReminderID, since)
	if err != nil {
		remLogger.Error("failed to check recent gifts", "error", err)
		return
	}

	occasionDetails := fmt.Sprintf("%s em %s", rem.Type, occDateTrunc.Format("2006-01-02"))
	if rem.Message != "" {
		occasionDetails += "; " + rem.Message
	}

	if len(recentGifts) > 0 {
		remLogger.Info("recent gifts exist for this reminder (notifications=P2), skipping generation")
		if err := j.jobLogRepo.Insert(ctx, rem.ReminderID, occDateTrunc); err != nil {
			remLogger.Error("failed to insert job log", "error", err)
		}
		return
	}

	remLogger.Info("generating suggestions")
	_, callErr := j.llmClient.SuggestionCreate(ctx, "", rem.FriendID, map[string]interface{}{
		"occasionDetails": occasionDetails,
		"reminderID":      rem.ReminderID,
		"source":          "automatic",
	})
	if callErr != nil {
		remLogger.Error("failed to generate suggestions", "error", callErr)
		return
	}

	if err := j.jobLogRepo.Insert(ctx, rem.ReminderID, occDateTrunc); err != nil {
		remLogger.Error("failed to insert job log", "error", err)
	}

	remLogger.Info("suggestions generated successfully")
}
