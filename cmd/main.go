package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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
	viewport        viewport.Model
	width           int
	height          int
}

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Quit       key.Binding
	Tab        key.Binding
	Enter      key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Enter, k.ScrollDown, k.ScrollUp, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Tab, k.Enter, k.Quit},
		{k.ScrollUp, k.ScrollDown},
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
		colWidth := tableWidth - 15 // Keep first column fixed
		m.containersTable.SetColumns([]table.Column{
			{Title: "Containers", Width: 10},
			{Title: "Image", Width: colWidth},
		})
		m.imageTable.SetColumns([]table.Column{
			{Title: "Images", Width: 10},
			{Title: "Tag", Width: colWidth},
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
					builder.WriteString(fmt.Sprintf("%s\n%s\n", titleStyle.Render("Networks:"), strings.Join(containerDetails.Networks, "\n")))
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
	tables := lipgloss.JoinVertical(lipgloss.Left,
		baseStyle.Render(m.containersTable.View()),
		baseStyle.Render(m.imageTable.View()),
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

	viewport := viewport.New(80, 20)
	viewport.SetContent("Select an item to see details")

	m := model{
		keys:            keys,
		help:            help,
		containersTable: containersTable,
		containers:      containers,
		imageTable:      imageTable,
		images:          images,
		viewport:        viewport,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
