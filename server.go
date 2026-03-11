package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var serverLogCh chan<- string

func startServer(cfg *Config, logCh chan<- string) error {
	serverLogCh = logCh

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": version})
	})

	mux.HandleFunc("/backups/", func(w http.ResponseWriter, r *http.Request) {
		if !auth(cfg.Token, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			logLine("✗", "auth", r.RemoteAddr+" — bad token")
			return
		}
		job := strings.TrimPrefix(r.URL.Path, "/backups/")
		job = strings.TrimSuffix(job, "/")
		if job == "" {
			listAllHandler(cfg, w, r)
			return
		}
		switch r.Method {
		case http.MethodPost:
			uploadHandler(cfg, job, w, r)
		case http.MethodGet:
			listHandler(cfg, job, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}
	logLine("●", "server", fmt.Sprintf("listening on %s", addr))
	return srv.ListenAndServe()
}

func auth(token string, r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer "+token
}

func uploadHandler(cfg *Config, job string, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	name, err := saveSnapshot(cfg.StoragePath, job, data)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		logLine("✗", job, fmt.Sprintf("storage error: %v", err))
		return
	}
	size := formatSize(int64(len(data)))
	logLine("↑", job, fmt.Sprintf("%s  (%s)", name, size))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"snapshot": name, "size": size})
}

func listHandler(cfg *Config, job string, w http.ResponseWriter, r *http.Request) {
	snaps, err := listSnapshots(cfg.StoragePath, job)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snaps)
}

func listAllHandler(cfg *Config, w http.ResponseWriter, r *http.Request) {
	jobs, err := listJobs(cfg.StoragePath)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func logLine(symbol, job, msg string) {
	line := fmt.Sprintf("  %s  %-16s  %s  %s", symbol, job, time.Now().Format("15:04:05"), msg)
	if serverLogCh != nil {
		serverLogCh <- line
	} else {
		fmt.Println(line)
	}
}
