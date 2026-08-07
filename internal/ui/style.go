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

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#336791")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#336791")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#336791")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)
