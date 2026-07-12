package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func snapshotDir(storagePath, job string) string {
	return filepath.Join(storagePath, job)
}

func isSnapshotFile(name string) bool {
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tar.gz.age")
}

// saveSnapshotStream streams an upload to disk without buffering it in
// memory. It writes to a temp file first so interrupted uploads never leave
// a half-written snapshot behind.
func saveSnapshotStream(storagePath, job, ext string, r io.Reader) (string, int64, error) {
	dir := snapshotDir(storagePath, job)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, err
	}

	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return "", 0, err
	}
	size, err := io.Copy(tmp, r)
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp.Name())
		return "", 0, err
	}

	// pick a free name; two uploads in the same second get -2, -3, …
	base := time.Now().Format("2006-01-02_15-04-05")
	name := base + ext
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
			break
		}
		name = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		os.Remove(tmp.Name())
		return "", 0, err
	}
	return name, size, nil
}

type SnapshotEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func listSnapshots(storagePath, job string) ([]SnapshotEntry, error) {
	dir := snapshotDir(storagePath, job)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []SnapshotEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var snaps []SnapshotEntry
	for _, e := range entries {
		if isSnapshotFile(e.Name()) {
			var size int64
			if info, err2 := e.Info(); err2 == nil {
				size = info.Size()
			}
			snaps = append(snaps, SnapshotEntry{Name: e.Name(), Size: size})
		}
	}
	sort.Slice(snaps, func(i, j int) bool { return snaps[i].Name > snaps[j].Name })
	return snaps, nil
}

func pruneSnapshotsServer(storagePath, job string, keep int) (int, error) {
	dir := snapshotDir(storagePath, job)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var snaps []string
	for _, e := range entries {
		if isSnapshotFile(e.Name()) {
			snaps = append(snaps, e.Name())
		}
	}
	sort.Strings(snaps) // oldest first
	deleted := 0
	for len(snaps) > keep {
		if err := os.Remove(filepath.Join(dir, snaps[0])); err != nil {
			return deleted, err
		}
		snaps = snaps[1:]
		deleted++
	}
	return deleted, nil
}

func listJobs(storagePath string) ([]string, error) {
	entries, err := os.ReadDir(storagePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var jobs []string
	for _, e := range entries {
		if e.IsDir() {
			jobs = append(jobs, e.Name())
		}
	}
	return jobs, nil
}

func readLogLines(n int) []string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".zipp-nest", "server.log")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
