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
	var tsItem string
	if !m.tsStatus.installed {
		tsItem = "Setup Tailscale"
	} else if m.tsStatus.running {
		tsItem = "Disconnect Tailscale"
	} else {
		tsItem = "Connect Tailscale"
	}
	if m.srvStatus.running {
		return []string{"Stop server", "Connection info", tsItem, "Quit"}
	}
	return []string{"Start server", "Connection info", tsItem, "Quit"}
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
		switch items[m.cursor] {
		case "Start server":
			if !m.tsStatus.running {
				m.page = pageResult
				m.resultLines = []string{
					"",
					styleError.Render("  ✗ Tailscale is not connected"),
					"",
					styleDim.Render("  select \"Setup Tailscale\" first"),
				}
				return m, nil
			}
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  starting server...")}
			return m, startServiceCmd()

		case "Stop server":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  stopping server...")}
			return m, stopServiceCmd()

		case "Setup Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  installing Tailscale...")}
			return m, installTailscaleCmd()

		case "Connect Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  connecting to Tailscale...")}
			return m, tailscaleUpCmd()

		case "Disconnect Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  disconnecting from Tailscale...")}
			return m, tailscaleDownCmd()

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
