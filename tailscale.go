package main

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type tailscaleStatus struct {
	installed bool
	running   bool
	ip        string
}

func checkTailscale() tailscaleStatus {
	if _, err := exec.LookPath("tailscale"); err != nil {
		return tailscaleStatus{}
	}
	out, err := exec.Command("tailscale", "ip", "-4").Output()
	if err != nil {
		return tailscaleStatus{installed: true, running: false}
	}
	ip := strings.TrimSpace(string(out))
	return tailscaleStatus{installed: true, running: ip != "", ip: ip}
}

type tailscaleCheckMsg tailscaleStatus

func checkTailscaleCmd() tea.Cmd {
	return func() tea.Msg {
		return tailscaleCheckMsg(checkTailscale())
	}
}

type tailscaleDoneMsg struct{ err error }

func installTailscaleCmd() tea.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("bash", "-c", "curl -fsSL https://tailscale.com/install.sh | sh && tailscale up")
	} else {
		// macOS: try brew first
		cmd = exec.Command("bash", "-c", "brew install tailscale && brew services start tailscale && tailscale up")
	}
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}
