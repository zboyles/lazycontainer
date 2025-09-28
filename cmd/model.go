package main

import (
	"fmt"
	"strings"
	"time"

	"lazycontainer/pkg/container"
	"lazycontainer/pkg/image"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	keys            keyMap
	help            help.Model
	containersTable table.Model
	containers      []container.Container
	imageTable      table.Model
	images          []image.Image
	viewport        viewport.Model
	width           int
	height          int
	popup           popupModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Measure help height and reserve a small bottom padding so the legend
		// doesn't touch the terminal border.
		helpRawHeight := lipgloss.Height(m.help.View(m.keys))
		helpPadding := 1
		mainContentHeight := m.height - helpRawHeight - helpPadding

		// Add a one-line gap between the two left tables. Each table's visual box
		// adds 2 lines (top+bottom) due to borders, so subtract 4 for the two
		// tables plus the 1-line gap before splitting the remainder evenly.
		gap := 1
		tableHeight := (mainContentHeight - 4 - gap) / 2
		topTableHeight := tableHeight
		bottomTableHeight := tableHeight
		m.containersTable.SetHeight(topTableHeight)
		m.imageTable.SetHeight(bottomTableHeight)

		// The viewport itself is wrapped with a border via baseStyle in View(), so
		// give it an inner height that's 2 lines less than the available area.
		m.viewport.Height = mainContentHeight - 2
		m.viewport.Width = m.width/2 - 5

		// Popup sizing (centered modal)
		m.popup.SetSize(m.width, mainContentHeight)

		tableWidth := m.width / 2
		m.containersTable.SetWidth(tableWidth)
		m.imageTable.SetWidth(tableWidth)

		// Recalculate column widths
		containerImageColWidth := tableWidth - 15 // Keep first column fixed
		m.containersTable.SetColumns([]table.Column{
			{Title: "Containers", Width: 10},
			{Title: "Image", Width: containerImageColWidth},
		})

		imageTagColWidth := tableWidth / 3
		imageNameColWidth := tableWidth - imageTagColWidth - 5 // Adjust for padding/borders
		m.imageTable.SetColumns([]table.Column{
			{Title: "Images", Width: imageNameColWidth},
			{Title: "Tag", Width: imageTagColWidth},
		})

	case tea.KeyMsg:
		// If popup is active, let it handle keys first
		if m.popup.active {
			var cmd tea.Cmd
			m.popup, cmd = m.popup.Update(msg)
			return m, cmd
		}
		if key.Matches(msg, m.keys.Enter) {
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			var content string

			if m.containersTable.Focused() {
				index := m.containersTable.Cursor()
				containerSelected := m.containers[index]
				containerDetails, err := container.GetDetails(containerSelected.ID)
				if err != nil {
					content = fmt.Sprintf("Error inspecting container %s: %v", containerSelected.ID, err)
				} else {
					var builder strings.Builder
					builder.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("ID:"), containerDetails.ID))
					builder.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("Image:"), containerDetails.Image))
					builder.WriteString(fmt.Sprintf("%s %d\n", titleStyle.Render("CPU:"), containerDetails.CPU))
					builder.WriteString(fmt.Sprintf("%s %d\n", titleStyle.Render("Memory:"), containerDetails.Memory))
					builder.WriteString(fmt.Sprintf("%s\n%s", titleStyle.Render("Networks:"), strings.Join(containerDetails.Networks, "\n")))
					builder.WriteString(fmt.Sprintf("%s\n%s", titleStyle.Render("Environment:"), strings.Join(containerDetails.Environment, "\n")))
					content = builder.String()
				}
			}

			if m.imageTable.Focused() {
				imageDetails, err := image.GetDetails(m.imageTable.SelectedRow()[0])
				if err != nil {
					content = fmt.Sprintf("Error inspecting image %s: %v", m.imageTable.SelectedRow()[0], err)
				} else {
					createdDataTime, _ := time.Parse(time.RFC3339, imageDetails.Created)
					localTime := createdDataTime.Local()
					formattedDateTime := localTime.Format("Mon, 02 Jan 2006 15:04:05 -07")
					sizeMB := float64(imageDetails.Size) / (1024 * 1024)

					var builder strings.Builder
					builder.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("Name:"), imageDetails.Name))
					builder.WriteString(fmt.Sprintf("%s %s\n", titleStyle.Render("ID:"), imageDetails.Id))
					builder.WriteString(fmt.Sprintf("%s %.2fMB\n", titleStyle.Render("Size:"), sizeMB))
					builder.WriteString(fmt.Sprintf("%s %s", titleStyle.Render("Created:"), formattedDateTime))
					content = builder.String()
				}
			}
			m.viewport.SetContent(content)
			m.viewport.GotoTop()
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			if m.containersTable.Focused() {
				m.containersTable.Blur()
				m.imageTable.Focus()
			} else {
				m.imageTable.Blur()
				m.containersTable.Focus()
			}
		case key.Matches(msg, m.keys.ScrollUp):
			m.viewport.LineUp(1)
		case key.Matches(msg, m.keys.ScrollDown):
			m.viewport.LineDown(1)
		case key.Matches(msg, m.keys.Actions):
			// Build context menu items based on current focus
			items := m.contextMenuItems()
			m.popup.OpenContext(items)
		}
	}

	if m.containersTable.Focused() {
		m.containersTable, cmd = m.containersTable.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.imageTable, cmd = m.imageTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Handle custom messages that set viewport content
	switch v := msg.(type) {
	case viewportMsg:
		m.viewport.SetContent(v.content)
		m.viewport.GotoTop()
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var containerStyle, imageStyle lipgloss.Style
	if m.containersTable.Focused() {
		containerStyle = focusedStyle
		imageStyle = baseStyle
	} else {
		containerStyle = baseStyle
		imageStyle = focusedStyle
	}

	// Add a visual one-line spacer between the two tables without borders.
	spacer := lipgloss.NewStyle().Height(1).Render("")
	tables := lipgloss.JoinVertical(lipgloss.Left,
		containerStyle.Render(m.containersTable.View()),
		spacer,
		imageStyle.Render(m.imageTable.View()),
	)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		tables,
		baseStyle.Render(m.viewport.View()),
	)

	helpView := helpStyle.Render(m.help.View(m.keys))

	root := lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		helpView,
	)

	if m.popup.active {
		// Draw only the modal centered in the screen while active.
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.popup.View())
	}

	return root
}

// contextMenuItems builds the actions available for the current selection.
func (m model) contextMenuItems() []actionItem {
	var items []actionItem
	if m.containersTable.Focused() && len(m.containers) > 0 {
		idx := m.containersTable.Cursor()
		if idx >= 0 && idx < len(m.containers) {
			c := m.containers[idx]
			items = append(items,
				actionItem{title: "Inspect", desc: "Show container details", enabled: true, onSelect: func() tea.Cmd {
					return func() tea.Msg { return tea.KeyMsg{Type: tea.KeyEnter} }
				}},
				actionItem{title: "Logs", desc: "Open logs in the right panel", enabled: true, onSelect: func() tea.Cmd {
					return func() tea.Msg {
						content, err := container.GetLogs(c.ID)
						if err != nil {
							content = fmt.Sprintf("Error: %v", err)
						}
						return viewportMsg{content: content}
					}
				}},
				actionItem{title: "Start", desc: "Start container", enabled: false},
				actionItem{title: "Stop", desc: "Stop container", enabled: false},
				actionItem{title: "Remove", desc: "Remove container", enabled: false},
			)
		}
	} else if m.imageTable.Focused() && len(m.images) > 0 {
		row := m.imageTable.SelectedRow()
		name := ""
		if len(row) > 0 {
			name = row[0]
		}
		items = append(items,
			actionItem{title: "Inspect", desc: "Show image details", enabled: true, onSelect: func() tea.Cmd {
				return func() tea.Msg { return tea.KeyMsg{Type: tea.KeyEnter} }
			}},
			actionItem{title: "Pull", desc: "Pull image", enabled: false},
			actionItem{title: "Run", desc: "Run image", enabled: false},
			actionItem{title: "Remove", desc: "Remove image", enabled: false},
		)
		_ = name
	}
	if len(items) == 0 {
		items = []actionItem{{title: "No actions available", desc: "", enabled: false}}
	}
	return items
}

// viewportMsg is a small message to set content in the viewport.
type viewportMsg struct{ content string }
