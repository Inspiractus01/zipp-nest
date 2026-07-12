package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveSnapshotStream(t *testing.T) {
	dir := t.TempDir()
	name, size, err := saveSnapshotStream(dir, "docs", ".tar.gz.age", strings.NewReader("payload"))
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len("payload")) {
		t.Errorf("size = %d", size)
	}
	if !strings.HasSuffix(name, ".tar.gz.age") {
		t.Errorf("name = %q, want .tar.gz.age suffix", name)
	}
	data, err := os.ReadFile(filepath.Join(dir, "docs", name))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "payload" {
		t.Errorf("content = %q", data)
	}
	// no leftover temp files
	entries, _ := os.ReadDir(filepath.Join(dir, "docs"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".upload-") {
			t.Errorf("leftover temp file %s", e.Name())
		}
	}
}

func TestSaveSnapshotStreamCollision(t *testing.T) {
	dir := t.TempDir()
	n1, _, err := saveSnapshotStream(dir, "docs", ".tar.gz", strings.NewReader("a"))
	if err != nil {
		t.Fatal(err)
	}
	n2, _, err := saveSnapshotStream(dir, "docs", ".tar.gz", strings.NewReader("b"))
	if err != nil {
		t.Fatal(err)
	}
	if n1 == n2 {
		t.Errorf("same-second uploads must not overwrite: %q == %q", n1, n2)
	}
}

func TestListAndPruneMixedExtensions(t *testing.T) {
	dir := t.TempDir()
	job := filepath.Join(dir, "docs")
	os.MkdirAll(job, 0755)
	files := []string{
		"2026-01-01_00-00-00.tar.gz",
		"2026-01-02_00-00-00.tar.gz.age",
		"2026-01-03_00-00-00.tar.gz.age",
	}
	for _, f := range files {
		os.WriteFile(filepath.Join(job, f), []byte("x"), 0644)
	}
	os.WriteFile(filepath.Join(job, "readme.txt"), []byte("x"), 0644)

	snaps, err := listSnapshots(dir, "docs")
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) != 3 {
		t.Fatalf("listed %d snapshots, want 3", len(snaps))
	}
	if snaps[0].Name != "2026-01-03_00-00-00.tar.gz.age" {
		t.Errorf("newest first, got %q", snaps[0].Name)
	}

	deleted, err := pruneSnapshotsServer(dir, "docs", 1)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Errorf("deleted = %d, want 2", deleted)
	}
	if _, err := os.Stat(filepath.Join(job, "readme.txt")); err != nil {
		t.Error("unrelated file must survive pruning")
	}
	if _, err := os.Stat(filepath.Join(job, "2026-01-03_00-00-00.tar.gz.age")); err != nil {
		t.Error("newest snapshot must survive pruning")
	}
}

func TestValidJobName(t *testing.T) {
	for _, bad := range []string{"", ".", "..", "a/b", `a\b`} {
		if validJobName(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
	if !validJobName("my-docs") {
		t.Error("normal name should pass")
	}
}

func TestNewerThan(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.10.0", "1.9.0", true},
		{"1.9.0", "1.10.0", false},
		{"2.0.0", "1.99.99", true},
		{"1.0.0", "1.0.0", false},
	}
	for _, c := range cases {
		if newerThan(c.a, c.b) != c.want {
			t.Errorf("newerThan(%q, %q) != %v", c.a, c.b, c.want)
		}
	}
}
