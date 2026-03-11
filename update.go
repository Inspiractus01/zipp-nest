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

func newerThan(a, b string) bool {
	return fmt.Sprintf("%010s", a) > fmt.Sprintf("%010s", b)
}

func runUpdateCmd() tea.Cmd {
	cmd := exec.Command("bash", "-c",
		"curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash",
	)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return updateDoneMsg{err: err}
	})
}
