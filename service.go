package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type serverStatus struct {
	running bool
	method  string // "launchd", "systemd", "none"
}

func checkServerService() serverStatus {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("launchctl", "list", "com.zipp-nest.server").Output()
		if err == nil && !strings.Contains(string(out), "Could not find") {
			return serverStatus{running: true, method: "launchd"}
		}
	}
	if runtime.GOOS == "linux" {
		out, err := exec.Command("systemctl", "--user", "is-active", "zipp-nest").Output()
		if err == nil && strings.TrimSpace(string(out)) == "active" {
			return serverStatus{running: true, method: "systemd"}
		}
	}
	return serverStatus{running: false, method: "none"}
}

func installService(bin string) error {
	if runtime.GOOS == "darwin" {
		return installLaunchd(bin)
	}
	return installSystemd(bin)
}

func installLaunchd(bin string) error {
	home, _ := os.UserHomeDir()
	dir := home + "/Library/LaunchAgents"
	plist := dir + "/com.zipp-nest.server.plist"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.zipp-nest.server</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + bin + `</string>
		<string>serve</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`
	if err := os.WriteFile(plist, []byte(content), 0644); err != nil {
		return err
	}
	exec.Command("launchctl", "unload", plist).Run()
	return exec.Command("launchctl", "load", plist).Run()
}

func installSystemd(bin string) error {
	home, _ := os.UserHomeDir()
	dir := home + "/.config/systemd/user"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	unit := `[Unit]
Description=Zipp Nest backup server

[Service]
ExecStart=` + bin + ` serve
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`
	if err := os.WriteFile(dir+"/zipp-nest.service", []byte(unit), 0644); err != nil {
		return err
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return exec.Command("systemctl", "--user", "enable", "--now", "zipp-nest").Run()
}

func uninstallService() error {
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		plist := home + "/Library/LaunchAgents/com.zipp-nest.server.plist"
		exec.Command("launchctl", "unload", plist).Run()
		return os.Remove(plist)
	}
	exec.Command("systemctl", "--user", "disable", "--now", "zipp-nest").Run()
	home, _ := os.UserHomeDir()
	os.Remove(home + "/.config/systemd/user/zipp-nest.service")
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

// — tea cmds —

type serverStatusMsg serverStatus
type serviceActionDoneMsg struct{ err error }

func checkServerServiceCmd() tea.Cmd {
	return func() tea.Msg {
		return serverStatusMsg(checkServerService())
	}
}

func startServiceCmd() tea.Cmd {
	return func() tea.Msg {
		self, err := os.Executable()
		if err != nil {
			self = "zipp-nest"
		}
		err = installService(self)
		return serviceActionDoneMsg{err: err}
	}
}

func stopServiceCmd() tea.Cmd {
	return func() tea.Msg {
		err := uninstallService()
		return serviceActionDoneMsg{err: err}
	}
}

func stopServerAndDisconnectCmd() tea.Cmd {
	return func() tea.Msg {
		err := uninstallService()
		// best-effort tailscale down (works without sudo if operator is set)
		exec.Command("tailscale", "down").Run()
		return serviceActionDoneMsg{err: err}
	}
}
