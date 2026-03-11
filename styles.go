package main

import "github.com/charmbracelet/lipgloss"

var (
	colorTeal      = lipgloss.Color("#0d9488")
	colorTealLight = lipgloss.Color("#5eead4")
	colorMuted     = lipgloss.Color("#6b7280")
	colorGray      = lipgloss.Color("#4b5563")
	colorWhite     = lipgloss.Color("#e2e8f0")
	colorGreen     = lipgloss.Color("#86efac")
	colorRed       = lipgloss.Color("#f87171")
	colorYellow    = lipgloss.Color("#fbbf24")

	styleSelected = lipgloss.NewStyle().Foreground(colorTealLight).Bold(true)
	styleNormal   = lipgloss.NewStyle().Foreground(colorWhite)
	styleDim      = lipgloss.NewStyle().Foreground(colorMuted)
	styleSuccess  = lipgloss.NewStyle().Foreground(colorGreen)
	styleError    = lipgloss.NewStyle().Foreground(colorRed)
	styleWarning  = lipgloss.NewStyle().Foreground(colorYellow)
	styleLogo     = lipgloss.NewStyle().Foreground(colorTeal)
	styleAccent   = lipgloss.NewStyle().Foreground(colorTealLight).Bold(true)
	styleVersion  = lipgloss.NewStyle().Foreground(colorGray)
	styleHint     = lipgloss.NewStyle().Foreground(colorGray)
)

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
