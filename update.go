package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type updateResult struct {
	latest    string
	hasUpdate bool
}

type updateCheckMsg updateResult
type updateDoneMsg struct{ err error }

func checkForUpdate() updateResult {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Inspiractus01/zipp-nest/releases/latest")
	if err != nil {
		return updateResult{}
	}
	defer resp.Body.Close()

	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return updateResult{}
	}

	latest := strings.TrimPrefix(data.TagName, "v")
	if latest == "" || latest == version {
		return updateResult{latest: latest}
	}
	return updateResult{
		latest:    latest,
		hasUpdate: newerThan(latest, version),
	}
}

// semver compare: returns true if a > b (both "x.y.z")
func newerThan(a, b string) bool {
	pa := parseSemver(a)
	pb := parseSemver(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func parseSemver(v string) [3]int {
	var major, minor, patch int
	fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	return [3]int{major, minor, patch}
}

func runUpdateCmd() tea.Cmd {
	cmd := exec.Command("bash", "-c",
		"curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash",
	)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return updateDoneMsg{err: err}
	})
}
