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
	pageResult
)

type model struct {
	page         page
	cursor       int
	config       *Config
	tsStatus     tailscaleStatus
	srvStatus    serverStatus
	resultLines  []string
	resultErr    error
	animFrame    int
}

func newModel(cfg *Config) model {
	return model{config: cfg}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(checkTailscaleCmd(), checkServerServiceCmd())
}

func (m model) menuItems() []string {
	if m.srvStatus.running {
		return []string{"Stop server", "Connection info", "Setup Tailscale", "Quit"}
	}
	return []string{"Start server", "Connection info", "Setup Tailscale", "Quit"}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tailscaleCheckMsg:
		m.tsStatus = tailscaleStatus(msg)

	case serverStatusMsg:
		m.srvStatus = serverStatus(msg)

	case serviceActionDoneMsg:
		m.resultErr = msg.err
		return m, checkServerServiceCmd()

	case tailscaleDoneMsg:
		m.resultErr = msg.err
		m.page = pageMenu
		return m, checkTailscaleCmd()

	case tickMsg:
		m.animFrame++
		return m, tickCmd()

	case tea.KeyMsg:
		switch m.page {
		case pageMenu:
			return m.updateMenu(msg)
		case pageResult:
			switch msg.String() {
			case "enter", "esc", "q":
				m.page = pageMenu
				m.resultLines = nil
				m.resultErr = nil
			}
		}
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "enter", " ":
		selected := items[m.cursor]
		switch selected {
		case "Start server":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  starting server service...")}
			return m, startServiceCmd()
		case "Stop server":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  stopping server service...")}
			return m, stopServiceCmd()
		case "Setup Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  installing Tailscale...")}
			return m, installTailscaleCmd()
		case "Connection info":
			host := m.tsStatus.ip
			if host == "" {
				host = "your-server-ip"
			}
			addr := fmt.Sprintf("%s:%d", host, m.config.Port)
			connStr := fmt.Sprintf("zipp nest add %s %s", addr, m.config.Token)
			m.page = pageResult
			m.resultLines = []string{
				"",
				"  " + styleDim.Render("address  ") + styleNormal.Render(addr),
				"  " + styleDim.Render("token    ") + styleAccent.Render(m.config.Token),
				"",
				styleDim.Render("  run this on your other machine:"),
				"",
				"  " + styleSelected.Render(connStr),
			}
		case "Quit":
			return m, tea.Quit
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	switch m.page {
	case pageResult:
		return m.viewResult()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(renderHeader())
	b.WriteString("\n")

	// server status
	srvLine := "  Server     "
	if m.srvStatus.running {
		srvLine += styleSuccess.Render("● running") + styleDim.Render("  ("+m.srvStatus.method+")")
	} else {
		srvLine += styleDim.Render("○ stopped")
	}
	b.WriteString(srvLine + "\n")

	// tailscale status
	tsLine := "  Tailscale  "
	if m.tsStatus.installed && m.tsStatus.running {
		tsLine += styleSuccess.Render("● connected  ") + styleDim.Render(m.tsStatus.ip)
	} else if m.tsStatus.installed {
		tsLine += styleWarning.Render("○ not connected")
	} else {
		tsLine += styleError.Render("○ not installed")
	}
	b.WriteString(tsLine + "\n\n")

	items := m.menuItems()
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(styleSelected.Render("▸ "+item) + "\n")
		} else {
			b.WriteString(styleNormal.Render("  "+item) + "\n")
		}
	}

	b.WriteString(styleHint.Render("\n  ↑↓ navigate · enter select · q quit"))
	return b.String()
}

func (m model) viewResult() string {
	var b strings.Builder
	b.WriteString(renderHeader())
	b.WriteString("\n")

	buzz := []string{"bzz   ", " bzz  ", "  bzz ", "   bzz"}
	b.WriteString(styleDim.Render("  "+buzz[m.animFrame%len(buzz)]) + "\n\n")

	for _, line := range m.resultLines {
		b.WriteString(line + "\n")
	}

	if m.resultErr != nil {
		b.WriteString("\n" + styleError.Render("  error: "+m.resultErr.Error()) + "\n")
	} else if len(m.resultLines) > 0 {
		b.WriteString("\n" + styleSuccess.Render("  ✓ done") + "\n")
	}

	b.WriteString(styleHint.Render("\n  enter / esc  back"))
	return b.String()
}

// — cmds —

type tickMsg struct{}

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

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
