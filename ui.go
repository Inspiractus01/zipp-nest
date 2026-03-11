package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type page int

const (
	pageMenu page = iota
	pageServer
	pageSetup
)

type model struct {
	page      page
	cursor    int
	config    *Config
	tsStatus  tailscaleStatus
	serverLog []string
	logCh     chan string
	animFrame int
}

type serverLogMsg string
type tickMsg struct{}

var menuItems = []string{
	"Start server",
	"Setup Tailscale",
	"Show token",
	"Quit",
}

func newModel(cfg *Config) model {
	return model{
		config: cfg,
	}
}

func (m model) Init() tea.Cmd {
	return checkTailscaleCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tailscaleCheckMsg:
		m.tsStatus = tailscaleStatus(msg)
		return m, nil

	case tailscaleDoneMsg:
		m.page = pageMenu
		return m, checkTailscaleCmd()

	case serverLogMsg:
		m.serverLog = append(m.serverLog, string(msg))
		if len(m.serverLog) > 50 {
			m.serverLog = m.serverLog[len(m.serverLog)-50:]
		}
		return m, readLogCmd(m.logCh)

	case tickMsg:
		m.animFrame++
		if m.page == pageServer {
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		switch m.page {
		case pageMenu:
			return m.updateMenu(msg)
		case pageServer:
			return m.updateServer(msg)
		}
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
	case "enter", " ":
		switch menuItems[m.cursor] {
		case "Start server":
			m.page = pageServer
			m.serverLog = nil
			ch := make(chan string, 64)
			m.logCh = ch
			return m, tea.Batch(
				startServerCmd(m.config, ch),
				readLogCmd(ch),
				tickCmd(),
			)
		case "Setup Tailscale":
			m.page = pageSetup
			return m, installTailscaleCmd()
		case "Show token":
			m.serverLog = []string{"  token: " + m.config.Token}
			// just show briefly in server view
			m.page = pageServer
			return m, nil
		case "Quit":
			return m, tea.Quit
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateServer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.page = pageMenu
		m.serverLog = nil
		m.logCh = nil
	}
	return m, nil
}

func (m model) View() string {
	switch m.page {
	case pageServer, pageSetup:
		return m.viewServer()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(renderHeader())
	b.WriteString("\n")

	// tailscale status
	tsLine := "  Tailscale  "
	if m.tsStatus.installed && m.tsStatus.running {
		tsLine += styleSuccess.Render("● connected  ") + styleDim.Render(m.tsStatus.ip)
	} else if m.tsStatus.installed {
		tsLine += styleWarning.Render("○ not connected")
	} else {
		tsLine += styleError.Render("○ not installed")
	}
	b.WriteString(styleDim.Render(tsLine) + "\n")
	b.WriteString("\n")

	for i, item := range menuItems {
		if i == m.cursor {
			b.WriteString(styleSelected.Render("▸ " + item))
		} else {
			b.WriteString(styleNormal.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString(styleHint.Render("\n  ↑↓ navigate · enter select · q quit"))
	return b.String()
}

func (m model) viewServer() string {
	var b strings.Builder
	b.WriteString(renderHeader())
	b.WriteString("\n")

	buzzFrames := []string{"bzz   ", " bzz  ", "  bzz ", "   bzz"}
	if m.page == pageServer {
		b.WriteString(styleDim.Render("  " + buzzFrames[m.animFrame%len(buzzFrames)]) + "\n\n")
	}

	for _, line := range m.serverLog {
		b.WriteString(line + "\n")
	}

	if m.page == pageServer {
		b.WriteString(styleHint.Render("\n  q / esc  back to menu"))
	}
	return b.String()
}

// — cmds —

type serverErrMsg struct{ err error }

func startServerCmd(cfg *Config, ch chan string) tea.Cmd {
	return func() tea.Msg {
		err := startServer(cfg, ch)
		return serverErrMsg{err: err}
	}
}

func readLogCmd(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return serverLogMsg(line)
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func keyHint(key, label string, c lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(key) +
		styleDim.Render(" "+label)
}

func runUI(cfg *Config) error {
	p := tea.NewProgram(newModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// placeholder so fmt is used
var _ = fmt.Sprintf
