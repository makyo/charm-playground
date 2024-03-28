package main

import (
	"log"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	started bool
	field   [][]int
}

type tickMsg time.Time

func nextGeneration(m model) [][]int {
	next := make([][]int, len(m.field))
	for i := 0; i < len(m.field); i++ {
		next[i] = make([]int, len(m.field[0]))
	}
	for y, _ := range m.field {
		for x, _ := range m.field[y] {
			neighborCount := 0

			// To the left
			if y-1 >= 0 {
				if x-1 >= 0 {
					neighborCount += m.field[y-1][x-1]
				}
				neighborCount += m.field[y-1][x]
				if x+1 < len(m.field[y]) {
					neighborCount += m.field[y-1][x+1]
				}
			}

			// On the same column
			if x-1 >= 0 {
				neighborCount += m.field[y][x-1]
			}
			if x+1 < len(m.field[y]) {
				neighborCount += m.field[y][x+1]
			}

			// To the right
			if y+1 < len(m.field) {
				if x-1 >= 0 {
					neighborCount += m.field[y+1][x-1]
				}
				neighborCount += m.field[y+1][x]
				if x+1 < len(m.field[y]) {
					neighborCount += m.field[y+1][x+1]
				}
			}
			if m.field[y][x] == 0 {
				if neighborCount == 3 {
					next[y][x] = 1
				} else {
					next[y][x] = 0
				}
			} else if m.field[y][x] == 1 {
				if neighborCount < 2 || neighborCount > 3 {
					next[y][x] = 0
				} else {
					next[y][x] = 1
				}
			}
		}
	}
	return next
}

func generateField(m model) model {
	for y, _ := range m.field {
		for x, _ := range m.field[y] {
			if rand.Intn(5) == 0 {
				m.field[y][x] = 1
			}
		}
	}
	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlR:
			m = generateField(m)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.field = make([][]int, msg.Height-1)
		for i := 0; i < msg.Height-1; i++ {
			m.field[i] = make([]int, msg.Width-1)
		}
		if !m.started {
			m = generateField(m)
			m.started = true
		}
	case tickMsg:
		m.field = nextGeneration(m)
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	var frame string
	for _, row := range m.field {
		for _, col := range row {
			if col == 1 {
				frame += "•"
			} else {
				frame += " "
			}
		}
		frame += "\n"
	}
	return frame
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
