package main

import "github.com/charmbracelet/lipgloss"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var focusedStyle = baseStyle.Copy().
	BorderForeground(lipgloss.Color("229"))
