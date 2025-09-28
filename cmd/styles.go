package main

import "github.com/charmbracelet/lipgloss"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var focusedStyle = baseStyle.
	BorderForeground(lipgloss.Color("229"))

// Adds breathing room below the help legend so it isn't flush with
// the terminal's bottom edge.
var helpStyle = lipgloss.NewStyle().
	PaddingBottom(1)

// Modal container for popup menus or confirmations.
var modalStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

// Very light dim to keep background visible while a modal is open.
var dimBackgroundStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("235")) // subtle dark gray
