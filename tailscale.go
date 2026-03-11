package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type tailscaleStatus struct {
	installed bool
	loggedIn  bool
	running   bool
	ip        string
}

func checkTailscale() tailscaleStatus {
	if _, err := exec.LookPath("tailscale"); err != nil {
		return tailscaleStatus{}
	}
	// connected?
	out, err := exec.Command("tailscale", "ip", "-4").Output()
	if err == nil {
		ip := strings.TrimSpace(string(out))
		if ip != "" {
			return tailscaleStatus{installed: true, loggedIn: true, running: true, ip: ip}
		}
	}
	// installed but not connected — check if logged in
	statusOut, _ := exec.Command("tailscale", "status").CombinedOutput()
	loggedIn := !strings.Contains(string(statusOut), "Logged out")
	return tailscaleStatus{installed: true, loggedIn: loggedIn, running: false}
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
		cmd = exec.Command("bash", "-c", "curl -fsSL https://tailscale.com/install.sh | sh")
	} else {
		cmd = exec.Command("bash", "-c", "brew install tailscale && brew services start tailscale")
	}
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}

// tailscale login — interactive, opens browser for auth
func tailscaleLoginCmd() tea.Cmd {
	return tea.ExecProcess(exec.Command("tailscale", "login"), func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}

// tailscale logout — non-interactive, expires node key
func tailscaleLogoutCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tailscale", "logout").CombinedOutput()
		if err != nil {
			return tailscaleDoneMsg{err: fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)}
		}
		return tailscaleDoneMsg{}
	}
}

// tailscale up — connect (non-interactive if already logged in)
func tailscaleUpCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tailscale", "up").CombinedOutput()
		if err != nil {
			return tailscaleDoneMsg{err: fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)}
		}
		return tailscaleDoneMsg{}
	}
}

// tailscale down — disconnect, stay logged in
func tailscaleDownCmd() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("tailscale", "down").CombinedOutput()
		if err != nil {
			return tailscaleDoneMsg{err: fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)}
		}
		return tailscaleDoneMsg{}
	}
}
