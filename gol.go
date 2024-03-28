package main

import (
	"log"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	// Keep track of whether we have started in case we get a WindowSizeMsg and don't actually want to regenerate the field.
	started bool

	// Maintain the field on which the automata live
	width int
	field []int
}

type tickMsg time.Time

// nextGeneration evolves the field of automata one generation based on the rules of Conway's Game of Life.
func nextGeneration(m model) model {
	// Create a new field based on the existing one.
	next := make([]int, len(m.field))

	// Loop over the cells.
	for i, cell := range m.field {
		neighborCount := 0

		// Count the adjacent living cells on the row above.
		if i-m.width > 0 {
			if i%m.width >= 0 {
				neighborCount += m.field[i-m.width-1]
			}
			neighborCount += m.field[i-m.width]
			if (i+1)%m.width != 0 {
				neighborCount += m.field[i-m.width+1]
			}
		}

		// Count the adjacent cells to either side.
		if i%m.width != 0 {
			neighborCount += m.field[i-1]
		}
		if i < len(m.field)-1 && (i+1)%m.width != 0 {
			neighborCount += m.field[i+1]
		}

		// Count the adjacent cells on the row below.
		if i+m.width < len(m.field) {
			if i%m.width >= 0 {
				neighborCount += m.field[i+m.width-1]
			}
			neighborCount += m.field[i+m.width]
			if (i+1)%m.width != 0 {
				neighborCount += m.field[i+m.width+1]
			}
		}

		// Evolve the current cell by the following rules:
		//
		// 1. A dead cell becomes live if it's surrounded by exactly three living cells to represent breeding.
		// 2. A living cell dies of loneliness if it has 0 or 1 neighbors.
		// 3. A living cell dies of overcrowding if it has more than 3 neighbors.
		// 4. A living cell stays alive if it has 2 or 3 neighbors.
		next[i] = cell
		if cell == 0 && neighborCount == 3 {
			next[i] = 1
			continue
		}
		if cell == 1 && (neighborCount < 2 || neighborCount > 3) {
			next[i] = 0
		}
	}
	m.field = next
	return m
}

// generateField generates a random field of automata, where each cell has a 1 in 5 chance of being alive.
func generateField(m model) model {
	for i, _ := range m.field {
		if rand.Intn(5) == 0 {
			m.field[i] = 1
		}
	}
	return m
}

// tick updates the model every 1/10 second.
func tick() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model. Since this requires a first WindowSizeMsg, we just send a tickMsg.
func (m model) Init() tea.Cmd {
	return tick()
}

// Update updates the state of the model based on various types of messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Key press messages
	case tea.KeyMsg:
		switch msg.Type {

		// Quit on Escape or Ctrl+C
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		// Regenerate the field on Ctrl+R
		case tea.KeyCtrlR:
			m = generateField(m)
			return m, nil
		}

	// Window size messages — we receive one initially, and then again on every resize
	case tea.WindowSizeMsg:

		// Reset the field to the correct size
		m.field = make([]int, (msg.Width)*(msg.Height))
		m.width = msg.Width

		// If it hasn't been started, generate the field.
		// TODO: if it has, trim the existing field
		if !m.started {
			m = generateField(m)
			m.started = true
		} else {
			log.Println("Resizing not yet implemented")
			m = generateField(m)
		}

	// Tick messages
	case tickMsg:

		// Evolve the next generation
		m = nextGeneration(m)
		return m, tick()
	}
	return m, nil
}

// View builds the entire screen's worth of cells to be printed by returning a • for a living cell or a space for a dead cell.
func (m model) View() string {
	var frame string

	// Loop over rows...
	for i, cell := range m.field {
		if cell == 1 {
			frame += "•"
		} else {
			frame += " "
		}
		if i > 0 && i%m.width == 0 {
			frame += "\n"
		}
	}

	return frame
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
