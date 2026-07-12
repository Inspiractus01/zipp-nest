package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	base := time.Now().Format(snapshotTimeLayout)
	name, err := claimSnapshotName(dir, tmp.Name(), base, ext)
	if err != nil {
		os.Remove(tmp.Name())
		return "", 0, err
	}
	return name, size, nil
}

// claimSnapshotName atomically hard-links tmpPath to a free name in dir
// derived from base+ext, trying base+"-2"+ext, base+"-3"+ext, … on
// collision. os.Link fails with an already-exists error if the target name
// is taken, so — unlike a stat-then-rename loop — two concurrent uploads can
// never both believe the same name is free and clobber each other.
func claimSnapshotName(dir, tmpPath, base, ext string) (string, error) {
	name := base + ext
	for i := 2; ; i++ {
		err := os.Link(tmpPath, filepath.Join(dir, name))
		if err == nil {
			_ = os.Remove(tmpPath) // tmp is now just a linked-away duplicate
			return name, nil
		}
		if !os.IsExist(err) {
			return "", err
		}
		name = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
}

type SnapshotEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

const snapshotTimeLayout = "2006-01-02_15-04-05"

// parseSnapshotName extracts the timestamp and collision sequence number
// encoded in a snapshot filename, e.g. "2006-01-02_15-04-05.tar.gz" (seq 1)
// or "2006-01-02_15-04-05-2.tar.gz" (seq 2) for same-second collisions.
func parseSnapshotName(name string) (t time.Time, seq int, ok bool) {
	base := name
	switch {
	case strings.HasSuffix(base, ".tar.gz.age"):
		base = strings.TrimSuffix(base, ".tar.gz.age")
	case strings.HasSuffix(base, ".tar.gz"):
		base = strings.TrimSuffix(base, ".tar.gz")
	default:
		return time.Time{}, 0, false
	}

	if len(base) > len(snapshotTimeLayout) && base[len(snapshotTimeLayout)] == '-' {
		ts := base[:len(snapshotTimeLayout)]
		n, err := strconv.Atoi(base[len(snapshotTimeLayout)+1:])
		if err != nil {
			return time.Time{}, 0, false
		}
		parsed, err := time.Parse(snapshotTimeLayout, ts)
		if err != nil {
			return time.Time{}, 0, false
		}
		return parsed, n, true
	}

	parsed, err := time.Parse(snapshotTimeLayout, base)
	if err != nil {
		return time.Time{}, 0, false
	}
	return parsed, 1, true
}

// snapshotChronoLess reports whether snapshot a happened before snapshot b.
// Sorting by parsed timestamp (with collision sequence as tiebreaker)
// instead of raw filename string keeps ordering correct even though '-'
// (used before the collision suffix) sorts before '.' (used before the
// extension), which would otherwise put "…-2.tar.gz" before "….tar.gz" even
// though it was written later. Names that fail to parse fall back to a
// plain string comparison rather than panicking or reordering unrelatedly.
func snapshotChronoLess(a, b string) bool {
	ta, sa, oka := parseSnapshotName(a)
	tb, sb, okb := parseSnapshotName(b)
	if !oka || !okb {
		return a < b
	}
	if !ta.Equal(tb) {
		return ta.Before(tb)
	}
	return sa < sb
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
	sort.Slice(snaps, func(i, j int) bool { return snapshotChronoLess(snaps[j].Name, snaps[i].Name) }) // newest first
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
	sort.Slice(snaps, func(i, j int) bool { return snapshotChronoLess(snaps[i], snaps[j]) }) // oldest first
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
