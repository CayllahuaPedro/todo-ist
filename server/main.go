package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo-ist/api"
	"todo-ist/database"
)

func main() {
	db, err := database.InitDB("todo-ist.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()
	dbHandler := api.DBHandler{
		DB: db,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", dbHandler.GetTodos)
	mux.HandleFunc("GET /todos/{id}", dbHandler.GetTodo)
	mux.HandleFunc("POST /todos", dbHandler.CreateTodo)
	mux.HandleFunc("PATCH /todos/{id}", dbHandler.UpdateTodo)
	mux.HandleFunc("DELETE /todos/{id}", dbHandler.DeleteTodo)
	mux.Handle("GET /", http.FileServer(http.Dir("../app")))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("shutdown error:", err)
	}

}
