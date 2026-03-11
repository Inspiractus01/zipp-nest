package main

import (
	"net"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// getLocalIP returns the machine's outbound local IPv4 address.
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

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
	statusOut, _ := exec.Command("tailscale", "status").CombinedOutput()
	status := string(statusOut)
	if strings.Contains(status, "Logged out") {
		return tailscaleStatus{installed: true, loggedIn: false, running: false}
	}
	if strings.Contains(status, "stopped") || strings.Contains(status, "Stopped") {
		return tailscaleStatus{installed: true, loggedIn: true, running: false}
	}
	out, err := exec.Command("tailscale", "ip", "-4").Output()
	if err != nil {
		return tailscaleStatus{installed: true, loggedIn: true, running: false}
	}
	ip := strings.TrimSpace(string(out))
	if ip == "" {
		return tailscaleStatus{installed: true, loggedIn: true, running: false}
	}
	return tailscaleStatus{installed: true, loggedIn: true, running: true, ip: ip}
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
	return tea.ExecProcess(exec.Command("sudo", "tailscale", "login"), func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}

// tailscale logout — needs sudo
func tailscaleLogoutCmd() tea.Cmd {
	return tea.ExecProcess(exec.Command("sudo", "tailscale", "logout"), func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}

// tailscale up — connect (sudo needed on most Linux setups)
func tailscaleUpCmd() tea.Cmd {
	return tea.ExecProcess(exec.Command("sudo", "tailscale", "up"), func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}

// tailscale down — disconnect, stay logged in
func tailscaleDownCmd() tea.Cmd {
	return tea.ExecProcess(exec.Command("sudo", "tailscale", "down"), func(err error) tea.Msg {
		return tailscaleDoneMsg{err: err}
	})
}
