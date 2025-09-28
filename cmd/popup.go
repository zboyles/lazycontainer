package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type popupType int

const (
	popupNone popupType = iota
	popupContext
)

type actionItem struct {
	title    string
	desc     string
	enabled  bool
	onSelect func() tea.Cmd
}

func (a actionItem) Title() string       { return a.title }
func (a actionItem) Description() string { return a.desc }
func (a actionItem) FilterValue() string { return a.title }

type popupModel struct {
	active bool
	ptype  popupType
	list   list.Model
	width  int
	height int
}

func newPopup(width, height int) popupModel {
	l := list.New(nil, list.NewDefaultDelegate(), width, height)
	l.Title = "Actions"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = l.Styles.Title.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true)
	return popupModel{list: l, width: width, height: height}
}

func (p popupModel) View() string {
	if !p.active {
		return ""
	}
	// Center the modal using a fixed size; caller ensures width/height sensible.
	return modalStyle.
		Width(p.width).
		Height(p.height).
		Render(p.list.View())
}

func (p *popupModel) SetSize(totalW, totalH int) {
	// modal body size
	mw := min(60, int(float64(totalW)*0.7))
	mh := min(18, int(float64(totalH)*0.6))
	p.width, p.height = mw, mh
	p.list.SetSize(mw-4, mh-4) // account for modal padding
}

func (p *popupModel) OpenContext(items []actionItem) {
	var li []list.Item
	for _, it := range items {
		// Mark disabled items visibly
		t := it.title
		if !it.enabled {
			t = fmt.Sprintf("%s (disabled)", t)
		}
		li = append(li, actionItem{title: t, desc: it.desc, enabled: it.enabled, onSelect: it.onSelect})
	}
	p.list.SetItems(li)
	p.active = true
	p.ptype = popupContext
}

func (p *popupModel) Close() { p.active = false; p.ptype = popupNone }

func (p popupModel) Update(msg tea.Msg) (popupModel, tea.Cmd) {
	if !p.active {
		return p, nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			p.active = false
			p.ptype = popupNone
			return p, nil
		case "enter":
			if it, ok := p.list.SelectedItem().(actionItem); ok && it.enabled && it.onSelect != nil {
				p.active = false
				p.ptype = popupNone
				return p, it.onSelect()
			}
		}
	}
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
