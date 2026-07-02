// @title          Gift App API
// @version        1.0
// @description    API para sugestões de presentes personalizadas.
// @host           localhost:8080
// @BasePath       /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/gift-app/api/docs"
	"github.com/gift-app/api/internal/adapter/gcsstorage"
	httpadapter "github.com/gift-app/api/internal/adapter/http"
	"github.com/gift-app/api/internal/adapter/llmapi"
	"github.com/gift-app/api/internal/adapter/postgres"
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

	userHandler := httpadapter.NewUserHandler(postgres.NewUserRepository(pool))
	friendHandler := httpadapter.NewFriendHandler(postgres.NewFriendRepository(pool))
	profileHandler := httpadapter.NewProfileHandler(postgres.NewProfileRepository(pool))
	profilePhotoStorage, err := gcsstorage.NewSignedURLServiceFromEnv(context.Background())
	if err != nil {
		log.Fatalf("could not initialize GCS signed url service: %v", err)
	}
	profilePhotoHandler := httpadapter.NewProfilePhotoHandler(postgres.NewFriendRepository(pool), profilePhotoStorage)
	profileAgentHandler := httpadapter.NewProfileAgentHandler(
		newLLMClientFromEnv(),
		postgres.NewFriendRepository(pool),
	)
	suggestionAgentHandler := httpadapter.NewSuggestionAgentHandler(
		newLLMClientFromEnv(),
		postgres.NewGiftRepository(pool),
		postgres.NewFriendRepository(pool),
		postgres.NewReminderRepository(pool),
	)
	giftHandler := httpadapter.NewGiftHandler(
		postgres.NewGiftRepository(pool),
		postgres.NewFriendRepository(pool),
		postgres.NewReminderRepository(pool),
	)
	reminderHandler := httpadapter.NewReminderHandler(postgres.NewReminderRepository(pool))

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

	// Reminders
	mux.HandleFunc("PUT /users/{userId}/reminders", reminderHandler.Create)
	mux.HandleFunc("GET /users/{userId}/reminders", reminderHandler.ListByUserID)
	mux.HandleFunc("POST /reminders/{reminderId}", reminderHandler.Update)

	handler := httpadapter.LoggingMiddleware(mux)

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
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
