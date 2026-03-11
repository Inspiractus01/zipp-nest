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
	page        page
	cursor      int
	config      *Config
	tsStatus    tailscaleStatus
	srvStatus   serverStatus
	resultLines []string
	resultErr   error
	animFrame   int
}

type tickMsg struct{}

func newModel(cfg *Config) model {
	return model{config: cfg}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(checkTailscaleCmd(), checkServerServiceCmd(), tickCmd())
}

func (m model) menuItems() []string {
	if !m.tsStatus.installed {
		return []string{"Setup Tailscale", "Quit"}
	}
	if !m.tsStatus.loggedIn {
		return []string{"Login to Tailscale", "Quit"}
	}
	if !m.tsStatus.running {
		return []string{"Connect Tailscale", "Logout from Tailscale", "Quit"}
	}
	// Tailscale connected
	if m.srvStatus.running {
		return []string{"Stop server", "Connection info", "Logout from Tailscale", "Quit"}
	}
	return []string{"Start server", "Connection info", "Logout from Tailscale", "Quit"}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tailscaleCheckMsg:
		m.tsStatus = tailscaleStatus(msg)

	case serverStatusMsg:
		m.srvStatus = serverStatus(msg)

	case serviceActionDoneMsg:
		// return to menu immediately, re-check status in background
		m.page = pageMenu
		m.resultLines = nil
		if msg.err != nil {
			m.resultErr = msg.err
		} else {
			m.resultErr = nil
		}
		return m, tea.Batch(checkServerServiceCmd(), checkTailscaleCmd())

	case tailscaleDoneMsg:
		m.page = pageMenu
		m.resultLines = nil
		if msg.err != nil && !strings.Contains(msg.err.Error(), "already") {
			m.resultErr = msg.err
		} else {
			m.resultErr = nil
		}
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
		switch items[m.cursor] {
		case "Start server":
			m.resultErr = nil
			return m, startServiceCmd()

		case "Stop server":
			m.resultErr = nil
			return m, stopServerAndDisconnectCmd()

		case "Setup Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  installing Tailscale...")}
			return m, installTailscaleCmd()

		case "Login to Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  opening browser for login...")}
			return m, tailscaleLoginCmd()

		case "Connect Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  connecting...")}
			return m, tailscaleUpCmd()

		case "Logout from Tailscale":
			m.resultErr = nil
			return m, tailscaleLogoutCmd()

		case "Connection info":
			ip := m.tsStatus.ip
			if ip == "" {
				m.page = pageResult
				m.resultLines = []string{
					"",
					styleError.Render("  ✗ Tailscale not connected — no IP available"),
				}
				return m, nil
			}
			addr := fmt.Sprintf("%s:%d", ip, m.config.Port)
			code, codeErr := encodeNestCode(ip)
			m.page = pageResult
			lines := []string{
				"",
				styleDim.Render("  give this to zipp on your other machine:"),
				"",
			}
			if codeErr == nil {
				lines = append(lines,
					"  "+styleAccent.Render(code)+styleDim.Render("  ← short code"),
					"",
					styleDim.Render("  or full address:"),
					"  "+styleDim.Render(addr),
				)
			} else {
				lines = append(lines, "  "+styleAccent.Render(addr))
			}
			m.resultLines = lines

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
	if !m.tsStatus.installed {
		tsLine += styleError.Render("○ not installed")
	} else if m.tsStatus.running {
		tsLine += styleSuccess.Render("● connected  ") + styleDim.Render(m.tsStatus.ip)
	} else if m.tsStatus.loggedIn {
		tsLine += styleWarning.Render("○ logged in, not connected")
	} else {
		tsLine += styleWarning.Render("○ logged out")
	}
	b.WriteString(tsLine + "\n\n")

	if m.resultErr != nil {
		b.WriteString("  " + styleError.Render("✗ "+m.resultErr.Error()) + "\n\n")
	}

	items := m.menuItems()
	for i, item := range items {
		if i == m.cursor {
			b.WriteString("  " + styleSelected.Render("▸ "+item) + "\n")
		} else {
			b.WriteString("  " + styleNormal.Render("  "+item) + "\n")
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
