package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"

	"github.com/Legibl/go_final_project/database"
	"github.com/Legibl/go_final_project/handlers"
)

func main() {
	db, err := database.InitializeDB("scheduler.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	repository := database.NewRepository(db)

	startServer(repository)
}

func startServer(repository *database.Repository) {
	r := chi.NewRouter()

	directoryPath := "web"
	fileServer := http.FileServer(http.Dir(directoryPath))
	r.Handle("/*", fileServer)

	r.Get("/api/nextdate", handlers.NextDateHandler)
	r.Post("/api/task", handlers.TaskHandler)
	r.Get("/api/tasks", handlers.HandleTaskGet(repository))
	r.Put("/api/task", handlers.HandleTaskPut(repository))
	r.Delete("/api/task", handlers.HandleTaskDelete(repository))
	r.Get("/api/task", handlers.HandleTaskID(repository))
	r.Post("/api/task/done", handlers.HandleTaskDone(repository))

	port := ":7540"
	fmt.Printf("Starting server on %s\n", port)
	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
