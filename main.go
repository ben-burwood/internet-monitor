package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"internet-monitor/internal/api"
	"internet-monitor/internal/config"
	"internet-monitor/internal/database"
	"internet-monitor/internal/scheduler"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()

	if err := database.Init(); err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer database.Close()

	sched := scheduler.New(cfg.Interval, cfg.Retention)
	go sched.Run(context.Background())

	webMux := http.NewServeMux()
	webMux.HandleFunc("GET /api/results", api.ListResults)
	webMux.HandleFunc("GET /api/results/latest", api.LatestResult)
	webMux.HandleFunc("GET /api/ip-history", api.IPHistory)
	webMux.HandleFunc("GET /api/servers", api.ListServers)
	webMux.HandleFunc("GET /api/settings", api.GetSettings)
	webMux.HandleFunc("PUT /api/settings", api.UpdateSettings)
	webMux.Handle("POST /api/run", api.RunNow(sched))

	// Serve Static Frontend
	webMux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	// Start web server on 8080
	log.Fatal(http.ListenAndServe("[::]:8080", webMux))
}
