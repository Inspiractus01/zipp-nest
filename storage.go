package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func snapshotDir(storagePath, job string) string {
	return filepath.Join(storagePath, job)
}

func saveSnapshot(storagePath, job string, data []byte) (string, error) {
	dir := snapshotDir(storagePath, job)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := time.Now().Format("2006-01-02_15-04-05") + ".tar.gz"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return name, nil
}

func listSnapshots(storagePath, job string) ([]string, error) {
	dir := snapshotDir(storagePath, job)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var snaps []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tar.gz") {
			snaps = append(snaps, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(snaps)))
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
		if strings.HasSuffix(e.Name(), ".tar.gz") {
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
