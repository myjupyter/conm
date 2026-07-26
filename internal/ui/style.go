package ui

import "charm.land/lipgloss/v2"

var pathStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#626262"))

var requiredStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF5F5F"))

var (
	highlightBg            = lipgloss.Color("#336791")
	highlightFg            = lipgloss.Color("#FFFFFF")
	highlightPlaceholderFg = lipgloss.Color("#CCD6E0")
)

var labelStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(highlightBg)

var markerStyle = lipgloss.NewStyle().
	Foreground(highlightFg).
	Background(highlightBg)
