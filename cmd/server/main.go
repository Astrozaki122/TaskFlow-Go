package main

import (
	"log"
	"net/http"
	"task-platform/internal/handler"
	"task-platform/internal/middleware"
	"task-platform/internal/view"
	"time"

	"task-platform/internal/config"
	"task-platform/internal/database"
)

func main() {

	config.Init()
	database.Connect()
	view.Init()

	mux := http.NewServeMux()
	registerRoutes(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func registerRoutes(mux *http.ServeMux) {

	// HEALTH CHECK
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(" Server is running"))
	})

	// AUTH
	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/login", handler.Login)

	// UI PAGES
	mux.HandleFunc("/login-page", handler.LoginPage)
	mux.HandleFunc("/dashboard", middleware.Auth(handler.Dashboard))

	// TASKS (PROTECTED)
	mux.HandleFunc("/tasks", middleware.Auth(handler.GetTasks))
	mux.HandleFunc("/tasks/create", middleware.Auth(handler.CreateTask))
	mux.HandleFunc("/tasks/delete", middleware.Auth(handler.DeleteTask))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}
