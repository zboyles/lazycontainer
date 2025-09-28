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

		helpViewHeight := lipgloss.Height(m.help.View(m.keys))
		mainContentHeight := m.height - helpViewHeight

		// Adjust for borders. Each table has a 2-line border. The viewport has one.
		tableHeight := (mainContentHeight - 4) / 2
		m.containersTable.SetHeight(tableHeight)
		m.imageTable.SetHeight(tableHeight)

		m.viewport.Height = mainContentHeight - 2
		m.viewport.Width = m.width/2 - 5

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

	tables := lipgloss.JoinVertical(lipgloss.Left,
		containerStyle.Render(m.containersTable.View()),
		imageStyle.Render(m.imageTable.View()),
	)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		tables,
		baseStyle.Render(m.viewport.View()),
	)

	helpView := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		helpView,
	)
}
