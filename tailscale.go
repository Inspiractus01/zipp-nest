package main

import (
	"encoding/json"
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
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return tailscaleStatus{installed: true}
	}
	var data struct {
		BackendState string `json:"BackendState"`
		Self         *struct {
			TailscaleIPs []string `json:"TailscaleIPs"`
		} `json:"Self"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return tailscaleStatus{installed: true}
	}
	switch data.BackendState {
	case "NeedsLogin", "NoState", "NeedsMachineAuth":
		return tailscaleStatus{installed: true, loggedIn: false, running: false}
	case "Stopped":
		return tailscaleStatus{installed: true, loggedIn: true, running: false}
	case "Running":
		ip := ""
		if data.Self != nil {
			for _, tsIP := range data.Self.TailscaleIPs {
				if !strings.Contains(tsIP, ":") {
					ip = tsIP
					break
				}
			}
		}
		return tailscaleStatus{installed: true, loggedIn: true, running: true, ip: ip}
	default:
		return tailscaleStatus{installed: true, loggedIn: false, running: false}
	}
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
