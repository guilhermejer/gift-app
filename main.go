// @title          Gift App API
// @version        1.0
// @description    API para sugestões de presentes personalizadas.
// @host           localhost:8080
// @BasePath       /
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/gift-app/api/docs"
	"github.com/gift-app/api/internal/adapter/gcsstorage"
	httpadapter "github.com/gift-app/api/internal/adapter/http"
	"github.com/gift-app/api/internal/adapter/llmapi"
	"github.com/gift-app/api/internal/adapter/postgres"
	"github.com/gift-app/api/internal/job"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	pool, err := postgres.NewPool(context.Background(), connStr)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	friendRepo := postgres.NewFriendRepository(pool)
	profileRepo := postgres.NewProfileRepository(pool)
	giftRepo := postgres.NewGiftRepository(pool)
	reminderRepo := postgres.NewReminderRepository(pool)
	jobLogRepo := postgres.NewSuggestionJobLogRepository(pool)

	llmClient := newLLMClientFromEnv()

	userHandler := httpadapter.NewUserHandler(userRepo)
	friendHandler := httpadapter.NewFriendHandler(friendRepo)
	profileHandler := httpadapter.NewProfileHandler(profileRepo)
	profilePhotoStorage, err := gcsstorage.NewSignedURLServiceFromEnv(context.Background())
	if err != nil {
		log.Fatalf("could not initialize GCS signed url service: %v", err)
	}
	profilePhotoHandler := httpadapter.NewProfilePhotoHandler(friendRepo, profilePhotoStorage)
	profileAgentHandler := httpadapter.NewProfileAgentHandler(llmClient, friendRepo)
	suggestionAgentHandler := httpadapter.NewSuggestionAgentHandler(llmClient, giftRepo, friendRepo, reminderRepo)
	giftHandler := httpadapter.NewGiftHandler(giftRepo, friendRepo, reminderRepo)
	reminderHandler := httpadapter.NewReminderHandler(reminderRepo)
	authHandler := httpadapter.NewAuthHandler(userRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Users
	mux.HandleFunc("PUT /users", userHandler.Create)
	mux.HandleFunc("GET /users/email", userHandler.GetByEmail)
	mux.HandleFunc("GET /users/{userId}", userHandler.GetByID)
	mux.HandleFunc("POST /users/{userId}", userHandler.Update)

	// Friends
	mux.HandleFunc("PUT /users/{userId}/friends", friendHandler.Create)
	mux.HandleFunc("GET /users/{userId}/friends", friendHandler.ListByUserID)
	mux.HandleFunc("GET /friends/{friendId}", friendHandler.GetByID)
	mux.HandleFunc("POST /friends/{friendId}", friendHandler.Update)

	// Profiles
	mux.HandleFunc("PUT /friends/{friendId}/profile", profileHandler.Save)
	mux.HandleFunc("GET /friends/{friendId}/profile", profileHandler.GetByFriendID)
	mux.HandleFunc("POST /friends/{friendId}/profile-photo/upload-url", profilePhotoHandler.CreateUploadURL)
	mux.HandleFunc("POST /friends/{friendId}/profile-photo/update-url", profilePhotoHandler.CreateUpdateURL)
	mux.HandleFunc("POST /friends/{friendId}/profile-photo/get-url", profilePhotoHandler.CreateGetURL)
	mux.HandleFunc("POST /friends/{friendId}/profile-photo/delete-url", profilePhotoHandler.CreateDeleteURL)
	mux.HandleFunc("POST /profiles/agent/chat", profileAgentHandler.Chat)
	mux.HandleFunc("POST /profiles/agent/finalize", profileAgentHandler.Finalize)
	mux.HandleFunc("DELETE /profiles/agent/session/{friendId}", profileAgentHandler.DeleteSession)
	mux.HandleFunc("POST /profiles/{friendId}/suggestions", suggestionAgentHandler.Create)
	mux.HandleFunc("POST /suggestions/agent/chat", suggestionAgentHandler.Chat)
	mux.HandleFunc("POST /suggestions/agent/finalize", suggestionAgentHandler.Finalize)

	// Gifts
	mux.HandleFunc("PUT /friends/{friendId}/gifts", giftHandler.Create)
	mux.HandleFunc("GET /friends/{friendId}/gifts", giftHandler.ListByFriendID)
	mux.HandleFunc("POST /gifts/{giftId}", giftHandler.Update)
	mux.HandleFunc("DELETE /gifts/{giftId}", giftHandler.Delete)

	// Reminders
	mux.HandleFunc("PUT /users/{userId}/reminders", reminderHandler.Create)
	mux.HandleFunc("GET /users/{userId}/reminders", reminderHandler.ListByUserID)
	mux.HandleFunc("POST /reminders/{reminderId}", reminderHandler.Update)
	mux.HandleFunc("DELETE /reminders/{reminderId}", reminderHandler.Delete)

	// Auth
	mux.HandleFunc("POST /auth/google", authHandler.Google)

	handler := httpadapter.LoggingMiddleware(mux)

	// Start suggestion job if enabled
	jobInterval := getJobInterval()
	jobEnabled := os.Getenv("GIFT_SUGGESTION_JOB_ENABLED") == "true"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if jobEnabled {
		suggestionJob := job.NewSuggestionJob(
			reminderRepo,
			userRepo,
			giftRepo,
			jobLogRepo,
			llmClient,
			jobInterval,
		)
		go suggestionJob.Run(ctx)
		slog.Info("suggestion job enabled", "interval", jobInterval)
	} else {
		slog.Info("suggestion job disabled")
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		slog.Info("server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for SIGINT
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	slog.Info("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
}

func getJobInterval() time.Duration {
	raw := os.Getenv("GIFT_SUGGESTION_JOB_INTERVAL_SECONDS")
	if raw == "" {
		return 24 * time.Hour
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(parsed) * time.Second
}

func newLLMClientFromEnv() *llmapi.Client {
	baseURL := os.Getenv("GIFT_LLM_API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	timeoutSeconds := 20
	if rawTimeout := os.Getenv("GIFT_LLM_API_TIMEOUT_SECONDS"); rawTimeout != "" {
		parsed, err := strconv.Atoi(rawTimeout)
		if err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	return llmapi.NewClient(baseURL, time.Duration(timeoutSeconds)*time.Second)
}
