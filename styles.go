package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette shared with the zipp client app, so the two TUIs read as one
// product family: a violet/lavender accent, gray/muted secondary tones, and
// green/red/yellow status colors. Each color adapts to the terminal's light
// or dark background — the values that used to be the only ones become the
// Dark side, and the Light side picks higher-contrast variants of the same
// hue for legibility on white/light terminal themes.
var (
	colorAccent = lipgloss.AdaptiveColor{Light: "#6d28d9", Dark: "#a78bfa"} // lavender
	colorLogo   = lipgloss.AdaptiveColor{Light: "#5b21b6", Dark: "#7b5ef8"} // violet
	colorMuted  = lipgloss.AdaptiveColor{Light: "#52525b", Dark: "#6b7280"}
	colorGray   = lipgloss.AdaptiveColor{Light: "#3f3f46", Dark: "#4b5563"}
	colorWhite  = lipgloss.AdaptiveColor{Light: "#1e293b", Dark: "#e2e8f0"}
	colorGreen  = lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#86efac"}
	colorRed    = lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f87171"}
	colorYellow = lipgloss.AdaptiveColor{Light: "#a16207", Dark: "#fbbf24"}

	styleSelected = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleNormal   = lipgloss.NewStyle().Foreground(colorWhite)
	styleDim      = lipgloss.NewStyle().Foreground(colorMuted)
	styleSuccess  = lipgloss.NewStyle().Foreground(colorGreen)
	styleError    = lipgloss.NewStyle().Foreground(colorRed)
	styleWarning  = lipgloss.NewStyle().Foreground(colorYellow)
	styleLogo     = lipgloss.NewStyle().Foreground(colorLogo)
	styleAccent   = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleVersion  = lipgloss.NewStyle().Foreground(colorGray)
	styleHint     = lipgloss.NewStyle().Foreground(colorGray)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorLogo).
			Padding(1, 2)
)

const (
	minPanelWidth     = 40
	maxPanelWidth     = 72
	defaultPanelWidth = 56
)

// clampInt keeps v within [min, max].
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// panelWidth derives a responsive content width for the bordered panel from
// the terminal's reported width (0 when unknown, e.g. before the first
// tea.WindowSizeMsg arrives), clamped to a sensible range so the layout
// never gets unreadably narrow or stretches absurdly wide.
func panelWidth(termWidth int) int {
	if termWidth <= 0 {
		return defaultPanelWidth
	}
	return clampInt(termWidth-8, minPanelWidth, maxPanelWidth)
}

// pageHeaderBar renders the one-line page identifier shown above the
// bordered panel on every screen, e.g. "ZIPP-NEST · SETTINGS".
func pageHeaderBar(name string) string {
	return styleAccent.Render("ZIPP-NEST · " + name)
}

// renderPage assembles a full screen: the page's header bar, its main
// content wrapped in a rounded, responsively-sized bordered panel, and a
// dim hint line below it.
func renderPage(windowWidth int, name, content, hint string) string {
	panel := stylePanel.Width(panelWidth(windowWidth)).Render(strings.TrimRight(content, "\n"))
	return pageHeaderBar(name) + "\n\n" + panel + "\n" + styleHint.Render(hint)
}

func renderHeader() string {
	nest := styleLogo.Render("  ,~~~~~,") + "\n" +
		styleLogo.Render(" (~") + styleAccent.Render("~~~~~") + styleLogo.Render("~)") + "\n" +
		styleLogo.Render("  `~~~~~`")

	name := styleAccent.Render("zipp-nest")
	ver := styleVersion.Render("v" + version)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		nest+"  ",
		"\n"+name+" "+ver,
	) + "\n"
}
