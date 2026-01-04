package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const filename = "tasks.json"

type mode int

const (
	viewMode mode = iota
	addMode
	editMode
)

type Model struct {
	tasks      []Task
	list       list.Model
	textInput  textinput.Model
	mode       mode
	selectedID int
	nextID     int
	err        error
}

func NewModel() Model {
	tasks, err := LoadTasks(filename)
	nextID := len(tasks) + 1

	for i := range tasks {
		if tasks[i].Priority == "" {
			tasks[i].Priority = "Low"
		}
		if tasks[i].ID >= nextID {
			nextID = tasks[i].ID + 1
		}
	}

	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = taskItem{task: t}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Task Tracker"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.Focus()

	return Model{
		tasks:     tasks,
		list:      l,
		textInput: ti,
		mode:      viewMode,
		nextID:    nextID,
		err:       err,
	}
}

type taskItem struct {
	task Task
}

func (i taskItem) Title() string {
	statusColor := lipgloss.Color("#FFFF00") // Yellow for Todo
	switch i.task.Status {
	case "In Progress":
		statusColor = lipgloss.Color("#00FFFF") // Cyan
	case "Done":
		statusColor = lipgloss.Color("#00FF00") // Green
	}
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	
	var priorityStr string
	var priorityStyle lipgloss.Style
	switch i.task.Priority {
	case "High":
		priorityStr = "[High]"
		priorityStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	case "Medium":
		priorityStr = "[Med]"
		priorityStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFF55"))
	case "Low":
		priorityStr = "[Low]"
		priorityStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888888"))
	default:
		priorityStr = ""
	}

	priorityRendered := priorityStyle.Render(priorityStr)

	title := i.task.Title
	if i.task.Complete {
		title = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("#888888")).Render(title)
	}

	return fmt.Sprintf("%s %s [%s]", priorityRendered, title, statusStyle.Render(i.task.Status))
}

func (i taskItem) Description() string { return "" }
func (i taskItem) FilterValue() string { return i.task.Title }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for help
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case viewMode:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				_ = SaveTasks(filename, m.tasks)
				return m, tea.Quit
			case "a":
				m.mode = addMode
				m.textInput.Reset()
				m.textInput.Focus()
				return m, textinput.Blink
			case "enter":
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					m.selectedID = selected.task.ID
					m.mode = editMode
					m.textInput.SetValue(selected.task.Title)
					m.textInput.Focus()
				}
				return m, textinput.Blink
			case "d": // Toggle done
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					for i := range m.tasks {
						if m.tasks[i].ID == selected.task.ID {
							m.tasks[i].Complete = !m.tasks[i].Complete
							if m.tasks[i].Complete {
								m.tasks[i].Status = "Done"
							} else {
								m.tasks[i].Status = "Todo"
							}
							m.list.SetItem(m.list.Index(), taskItem{task: m.tasks[i]})
							_ = SaveTasks(filename, m.tasks)
							break
						}
					}
				}
			case "p": // Set to In Progress
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					for i := range m.tasks {
						if m.tasks[i].ID == selected.task.ID {
							m.tasks[i].Status = "In Progress"
							m.list.SetItem(m.list.Index(), taskItem{task: m.tasks[i]})
							_ = SaveTasks(filename, m.tasks)
							break
						}
					}
				}
			case "1":
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					updateTaskPriority(&m, selected.task.ID, "High")
				}
			case "2":
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					updateTaskPriority(&m, selected.task.ID, "Medium")
				}
			case "3":
				if selected, ok := m.list.SelectedItem().(taskItem); ok {
					updateTaskPriority(&m, selected.task.ID, "Low")
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case addMode, editMode:
			switch msg.String() {
			case "esc":
				m.mode = viewMode
				return m, nil
			case "enter":
				title := m.textInput.Value()
				if title == "" {
					return m, nil
				}
				if m.mode == addMode {
					newTask := Task{ID: m.nextID, Title: title, Status: "Todo", Complete: false}
					m.tasks = append(m.tasks, newTask)
					m.list.InsertItem(len(m.list.Items()), taskItem{task: newTask})
					m.nextID++
				} else {
					for i := range m.tasks {
						if m.tasks[i].ID == m.selectedID {
							m.tasks[i].Title = title
							m.list.SetItem(m.list.Index(), taskItem{task: m.tasks[i]})
							break
						}
					}
				}
				_ = SaveTasks(filename, m.tasks)
				m.mode = viewMode
				return m, nil
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	switch m.mode {
	case addMode:
		return fmt.Sprintf("Add New Task:\n%s\n\n(esc to cancel)", m.textInput.View())
	case editMode:
		return fmt.Sprintf("Edit Task:\n%s\n\n(esc to cancel)", m.textInput.View())
	default:
		help := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("\n↑/↓ or j/k: navigate • a: add • enter: edit • p: in progress • d: toggle done • x: delete • 1/2/3: priority (High/Med/Low) • q/esc: quit")
		return m.list.View() + help
	}
}

func updateTaskPriority(m *Model, id int, priority string) {
	for i := range m.tasks {
		if m.tasks[i].ID == id {
			m.tasks[i].Priority = priority
			m.list.SetItem(i, taskItem{task: m.tasks[i]})
			_ = SaveTasks(filename, m.tasks)
			break
		}
	}
}