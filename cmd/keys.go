package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Quit       key.Binding
	Tab        key.Binding
	Enter      key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Actions    key.Binding
	PageLeft   key.Binding
	PageRight  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.ScrollDown, k.ScrollUp, k.PageLeft, k.PageRight, k.Actions, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Tab, k.Enter, k.Quit},
		{k.ScrollUp, k.ScrollDown},
		{k.PageLeft, k.PageRight},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("⇥", "switch focus"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↩", "select"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "scroll down"),
	),
	Actions: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "menu"),
	),
	PageLeft: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "page left"),
	),
	PageRight: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "page right"),
	),
}
