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
	httpadapter "github.com/gift-app/api/internal/adapter/http"
	"github.com/gift-app/api/internal/adapter/llmapi"
	"github.com/gift-app/api/internal/adapter/postgres"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://giftowner:giftowner@localhost:6432/giftdb?sslmode=disable"
	}

	pool, err := postgres.NewPool(context.Background(), connStr)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	userHandler := httpadapter.NewUserHandler(postgres.NewUserRepository(pool))
	friendHandler := httpadapter.NewFriendHandler(postgres.NewFriendRepository(pool))
	profileHandler := httpadapter.NewProfileHandler(postgres.NewProfileRepository(pool))
	profileAgentHandler := httpadapter.NewProfileAgentHandler(
		newLLMClientFromEnv(),
		postgres.NewFriendRepository(pool),
	)
	giftHandler := httpadapter.NewGiftHandler(postgres.NewGiftRepository(pool))
	reminderHandler := httpadapter.NewReminderHandler(postgres.NewReminderRepository(pool))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Users
	mux.HandleFunc("PUT /users", userHandler.Create)
	mux.HandleFunc("GET /users/{user_id}", userHandler.GetByID)
	mux.HandleFunc("POST /users/{user_id}", userHandler.Update)

	// Friends
	mux.HandleFunc("PUT /users/{user_id}/friends", friendHandler.Create)
	mux.HandleFunc("GET /users/{user_id}/friends", friendHandler.ListByUserID)
	mux.HandleFunc("GET /friends/{friend_id}", friendHandler.GetByID)
	mux.HandleFunc("POST /friends/{friend_id}", friendHandler.Update)

	// Profiles
	mux.HandleFunc("PUT /friends/{friend_id}/profile", profileHandler.Save)
	mux.HandleFunc("GET /friends/{friend_id}/profile", profileHandler.GetByFriendID)
	mux.HandleFunc("POST /profiles/agent/chat", profileAgentHandler.Chat)
	mux.HandleFunc("POST /profiles/agent/finalize", profileAgentHandler.Finalize)
	mux.HandleFunc("DELETE /profiles/agent/session/{session_id}", profileAgentHandler.DeleteSession)

	// Gifts
	mux.HandleFunc("PUT /friends/{friend_id}/gifts", giftHandler.Create)
	mux.HandleFunc("GET /friends/{friend_id}/gifts", giftHandler.ListByFriendID)
	mux.HandleFunc("POST /gifts/{gift_id}", giftHandler.Update)

	// Reminders
	mux.HandleFunc("PUT /users/{user_id}/reminders", reminderHandler.Create)
	mux.HandleFunc("GET /users/{user_id}/reminders", reminderHandler.ListByUserID)
	mux.HandleFunc("POST /reminders/{reminder_id}", reminderHandler.Update)

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
