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

	httpadapter "github.com/gift-app/api/internal/adapter/http"
	"github.com/gift-app/api/internal/adapter/postgres"
	_ "github.com/gift-app/api/docs"
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
	giftHandler := httpadapter.NewGiftHandler(postgres.NewGiftRepository(pool))
	reminderHandler := httpadapter.NewReminderHandler(postgres.NewReminderRepository(pool))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Users
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{user_id}", userHandler.GetByID)

	// Friends
	mux.HandleFunc("POST /users/{user_id}/friends", friendHandler.Create)
	mux.HandleFunc("GET /users/{user_id}/friends", friendHandler.ListByUserID)
	mux.HandleFunc("GET /friends/{friend_id}", friendHandler.GetByID)

	// Profiles
	mux.HandleFunc("PUT /friends/{friend_id}/profile", profileHandler.Save)
	mux.HandleFunc("GET /friends/{friend_id}/profile", profileHandler.GetByFriendID)

	// Gifts
	mux.HandleFunc("POST /friends/{friend_id}/gifts", giftHandler.Create)
	mux.HandleFunc("GET /friends/{friend_id}/gifts", giftHandler.ListByFriendID)

	// Reminders
	mux.HandleFunc("POST /users/{user_id}/reminders", reminderHandler.Create)
	mux.HandleFunc("GET /users/{user_id}/reminders", reminderHandler.ListByUserID)

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
