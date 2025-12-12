package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const filename = "tasks.json"

type mode IsNotExist

const (
	viewMode mode = iota
	addMode
	editMode
)

type Model struct {
	tasks		[]Task
	list		list.Model
	textInput	textinput.Model
	mode		mode
	selectedID	int
	nextID		int
	err			error
}