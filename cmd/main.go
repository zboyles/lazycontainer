package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	container "lazycontainer/pkg/container"
	image "lazycontainer/pkg/image"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	keys            keyMap
	help            help.Model
	containersTable table.Model
	containers      []container.Container
	imageTable      table.Model
	images          []image.Image
	infoBox         string
	width           int
	height          int
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Quit  key.Binding
	Tab   key.Binding
	Enter key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Tab, k.Enter, k.Quit},
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
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		// Recalculate table and info box dimensions
		m.containersTable.SetHeight(m.height/2 - 10)
		m.imageTable.SetHeight(m.height/2 - 10)
		m.containersTable.SetWidth(m.width/2 - 5)
		m.imageTable.SetWidth(m.width/2 - 5)

	case tea.KeyMsg:
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
		case key.Matches(msg, m.keys.Enter):
			if m.containersTable.Focused() {
				index := m.containersTable.Cursor()
				containerSelected := m.containers[index]
				containerDetails, err := container.GetDetails(containerSelected.ID)
				if err != nil {
					m.infoBox = fmt.Sprintf("Error inspecting container %s: %v", containerSelected.ID, err)
				} else {
					m.infoBox = fmt.Sprintf("ID: %s \nImage: %s \nCPU: %d \nMemory: %d \nNetworks: %s \nEnvironment: %s", containerDetails.ID, containerDetails.Image, containerDetails.CPU, containerDetails.Memory,
						lipgloss.JoinVertical(lipgloss.Left, containerDetails.Networks...),
						lipgloss.JoinVertical(lipgloss.Left, containerDetails.Environment...),
					)
				}
			}

			if m.imageTable.Focused() {
				imageDetails, err := image.GetDetails(m.imageTable.SelectedRow()[0])
				if err != nil {
					m.infoBox = fmt.Sprintf("Error inspecting image %s: %v", m.imageTable.SelectedRow()[0], err)
				} else {
					createdDataTime, _ := time.Parse(time.RFC3339, imageDetails.Created)
					localTime := createdDataTime.Local()
					formattedDateTime := localTime.Format("Mon, 02 Jan 2006 15:04:05 -07")
					sizeMB := float64(imageDetails.Size) / (1024 * 1024)
					m.infoBox = fmt.Sprintf("Name: %s \nID: %s \nSize: %.2fMB \nCreated: %s", imageDetails.Name, imageDetails.Id, sizeMB, formattedDateTime)
				}
			}
		}
	}

	if m.containersTable.Focused() {
		m.containersTable, cmd = m.containersTable.Update(msg)
	} else {
		m.imageTable, cmd = m.imageTable.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	tables := lipgloss.JoinVertical(lipgloss.Left,
		baseStyle.Render(m.containersTable.View()),
		baseStyle.Render(m.imageTable.View()),
	)

	infoBoxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width/2 - 5).
		Height(m.height - 12).
		Padding(1, 2)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
		tables,
		infoBoxStyle.Render(m.infoBox),
	)

	helpView := m.help.View(m.keys)

	return lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		helpView,
	)
}

func main() {
	// containers table
	containers, err := container.ListAll()
	if err != nil {
		fmt.Println("Error listing containers:", err)
	}

	containerRows := []table.Row{}
	for _, c := range containers {
		containerRows = append(containerRows, table.Row{c.State, c.Image})
	}

	containerColumns := []table.Column{
		{Title: "Containers", Width: 10},
		{Title: "", Width: 25},
	}

	containersTable := table.New(
		table.WithColumns(containerColumns),
		table.WithRows(containerRows),
		table.WithFocused(true),
		table.WithHeight(5),
	)

	styleContainers := table.DefaultStyles()
	styleContainers.Header = styleContainers.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	styleContainers.Selected = styleContainers.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	containersTable.SetStyles(styleContainers)

	// Images table
	images, err := image.ListAll()
	if err != nil {
		fmt.Println("Error listing images:", err)
	}

	imageRows := []table.Row{}
	for _, image := range images {
		imageRows = append(imageRows, table.Row{image.Name, image.Tag})
	}

	imageColumns := []table.Column{
		{Title: "Images", Width: 10},
		{Title: "", Width: 25},
	}

	imageTable := table.New(
		table.WithColumns(imageColumns),
		table.WithRows(imageRows),
		table.WithFocused(false),
		table.WithHeight(5),
	)

	styleImages := table.DefaultStyles()
	styleImages.Header = styleImages.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	styleImages.Selected = styleImages.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("201")).
		Bold(false)
	imageTable.SetStyles(styleImages)

	help := help.New()
	help.ShowAll = true

	m := model{
		keys:            keys,
		help:            help,
		containersTable: containersTable,
		containers:      containers,
		imageTable:      imageTable,
		images:          images,
		infoBox:         "Select an item to see details",
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
