package main

import "github.com/charmbracelet/lipgloss"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var focusedStyle = baseStyle.Copy().
	BorderForeground(lipgloss.Color("229"))

// Adds breathing room below the help legend so it isn't flush with
// the terminal's bottom edge.
var helpStyle = lipgloss.NewStyle().
	PaddingBottom(1)
