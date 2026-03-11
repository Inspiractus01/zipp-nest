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
	updateInfo  updateResult
	resultLines []string
	resultErr   error
	animFrame   int
}

type tickMsg struct{}
type periodicCheckMsg struct{}

func periodicCheckCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(time.Time) tea.Msg {
		return periodicCheckMsg{}
	})
}

func newModel(cfg *Config) model {
	return model{config: cfg}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		checkTailscaleCmd(),
		checkServerServiceCmd(),
		tickCmd(),
		periodicCheckCmd(),
		func() tea.Msg { return updateCheckMsg(checkForUpdate()) },
	)
}

func (m model) menuItems() []string {
	var items []string
	if m.srvStatus.running {
		items = append(items, "Stop server")
	} else {
		items = append(items, "Start server")
	}
	items = append(items, "Connection info")

	if !m.tsStatus.installed {
		items = append(items, "Setup Tailscale")
	} else if !m.tsStatus.loggedIn {
		items = append(items, "Login to Tailscale")
	} else {
		items = append(items, "Logout from Tailscale")
		if m.tsStatus.running {
			items = append(items, "Disable Tailscale")
		} else {
			items = append(items, "Enable Tailscale")
		}
	}
	if m.updateInfo.hasUpdate {
		items = append(items, "Run update")
	}
	items = append(items, "Quit")
	return items
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case updateCheckMsg:
		m.updateInfo = updateResult(msg)
		return m, nil

	case updateDoneMsg:
		if msg.err != nil {
			m.resultErr = msg.err
			m.page = pageMenu
			return m, nil
		}
		return m, tea.Quit

	case periodicCheckMsg:
		return m, tea.Batch(checkTailscaleCmd(), checkServerServiceCmd(), periodicCheckCmd())

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
			return m, stopServiceCmd()

		case "Setup Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  installing Tailscale...")}
			return m, installTailscaleCmd()

		case "Login to Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  opening browser for login...")}
			return m, tailscaleLoginCmd()

		case "Enable Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  connecting...")}
			return m, tailscaleUpCmd()

		case "Disable Tailscale":
			m.page = pageResult
			m.resultLines = []string{styleDim.Render("  disconnecting...")}
			return m, tailscaleDownCmd()

		case "Logout from Tailscale":
			m.resultErr = nil
			return m, tailscaleLogoutCmd()

		case "Connection info":
			lines := []string{""}
			addEntry := func(label, ip string) {
				addr := fmt.Sprintf("%s:%d", ip, m.config.Port)
				code, err := encodeNestCode(ip)
				lines = append(lines, styleDim.Render("  "+label+":"))
				if err == nil {
					lines = append(lines, "  "+styleAccent.Render(code)+styleDim.Render("  ← code"))
				}
				lines = append(lines, styleDim.Render("  "+addr), "")
			}
			if m.tsStatus.running {
				addEntry("tailscale", m.tsStatus.ip)
			}
			if localIP := getLocalIP(); localIP != "" {
				addEntry("local network", localIP)
			}
			if len(lines) == 1 {
				lines = append(lines, styleError.Render("  ✗ no network address available"))
			}
			m.page = pageResult
			m.resultLines = lines

		case "Run update":
			return m, runUpdateCmd()

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

	if m.updateInfo.hasUpdate {
		b.WriteString("  " + styleWarning.Render("● update available: v"+m.updateInfo.latest) + "\n\n")
	}

	if m.resultErr != nil {
		b.WriteString("  " + styleError.Render("✗ "+m.resultErr.Error()) + "\n\n")
	}

	items := m.menuItems()
	for i, item := range items {
		if i == m.cursor {
			b.WriteString("  " + styleSelected.Render("▸ "+item) + "\n")
		} else if item == "Run update" {
			b.WriteString("  " + styleWarning.Render("  "+item) + "\n")
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
