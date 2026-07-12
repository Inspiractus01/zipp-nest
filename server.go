package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var serverLogCh chan<- string

// validJobName rejects anything that could escape the storage directory.
func validJobName(job string) bool {
	return job != "" && job != "." && job != ".." &&
		!strings.ContainsAny(job, "/\\")
}

// authorized checks the bearer token in constant time.
func authorized(cfg *Config, r *http.Request) bool {
	if cfg.Token == "" {
		return true
	}
	got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return subtle.ConstantTimeCompare([]byte(got), []byte(cfg.Token)) == 1
}

func startServer(cfg *Config, logCh chan<- string) error {
	serverLogCh = logCh

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": version})
	})

	mux.HandleFunc("/backups/", func(w http.ResponseWriter, r *http.Request) {
		if !authorized(cfg, r) {
			logLine("✗", "auth", "rejected request from "+r.RemoteAddr)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/backups/")
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			listAllHandler(cfg, w, r)
			return
		}
		// split into job and optional snapshot name
		parts := strings.SplitN(path, "/", 2)
		job := parts[0]
		if !validJobName(job) {
			http.Error(w, "invalid job name", http.StatusBadRequest)
			return
		}
		if len(parts) == 2 && parts[1] != "" {
			if r.Method == http.MethodGet {
				downloadHandler(cfg, job, parts[1], w, r)
			} else {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodPost:
			uploadHandler(cfg, job, w, r)
		case http.MethodGet:
			listHandler(cfg, job, w, r)
		case http.MethodDelete:
			keep := 0
			fmt.Sscanf(r.URL.Query().Get("keep"), "%d", &keep)
			if keep <= 0 {
				http.Error(w, "invalid keep param", http.StatusBadRequest)
				return
			}
			deleted, err := pruneSnapshotsServer(cfg.StoragePath, job, keep)
			if err != nil {
				http.Error(w, "prune error", http.StatusInternalServerError)
				return
			}
			if deleted > 0 {
				logLine("✂", job, fmt.Sprintf("pruned %d old snapshot(s), keeping %d", deleted, keep))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"deleted": deleted})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}
	logLine("●", "server", fmt.Sprintf("listening on %s", addr))
	return srv.ListenAndServe()
}

func uploadHandler(cfg *Config, job string, w http.ResponseWriter, r *http.Request) {
	// clients mark age-encrypted uploads via header; keep the extension
	// so restores know whether to decrypt
	ext := ".tar.gz"
	if r.Header.Get("X-Zipp-Encrypted") == "age" {
		ext = ".tar.gz.age"
	}
	name, size, err := saveSnapshotStream(cfg.StoragePath, job, ext, r.Body)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		logLine("✗", job, fmt.Sprintf("storage error: %v", err))
		return
	}
	sizeStr := formatSize(size)
	logLine("↑", job, fmt.Sprintf("%s  (%s)", name, sizeStr))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"snapshot": name, "size": sizeStr})
}

func downloadHandler(cfg *Config, job, snapshot string, w http.ResponseWriter, r *http.Request) {
	// sanitize: no path traversal
	if strings.Contains(snapshot, "/") || strings.Contains(snapshot, "..") {
		http.Error(w, "invalid snapshot name", http.StatusBadRequest)
		return
	}
	path := filepath.Join(snapshotDir(cfg.StoragePath, job), snapshot)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, snapshot))
	if _, err := io.Copy(w, f); err != nil {
		// headers are already sent, so the client just sees a truncated
		// download; the best we can do here is log it for diagnosis.
		logLine("✗", job, fmt.Sprintf("download of %s failed partway: %v", snapshot, err))
		return
	}
	logLine("↓", job, snapshot)
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

func logFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zipp-nest", "server.log")
}

// maxLogFileSize caps how large server.log is allowed to grow before it's
// rotated; keeps a long-running server from filling the disk over time.
const maxLogFileSize = 5 * 1024 * 1024 // 5MB

// rotateLogIfNeeded moves the current log file to a single ".1" backup
// (overwriting any previous one) once it exceeds maxLogFileSize, so logLine
// always appends to a fresh file afterward. Kept intentionally simple: one
// rotation generation, no compression, no multiple backups.
func rotateLogIfNeeded(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < maxLogFileSize {
		return
	}
	os.Rename(path, path+".1")
}

func logLine(symbol, job, msg string) {
	line := fmt.Sprintf("  %s  %-16s  %s  %s", symbol, job, time.Now().Format("2006-01-02 15:04:05"), msg)
	if serverLogCh != nil {
		serverLogCh <- line
	} else {
		fmt.Println(line)
	}
	// always append to log file, rotating first if it's grown too large
	path := logFilePath()
	rotateLogIfNeeded(path)
	if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintln(f, line)
		f.Close()
	}
}
