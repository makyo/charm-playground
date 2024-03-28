package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	// Keep track of whether we have started in case we get a WindowSizeMsg and don't actually want to regenerate the field.
	started bool

	// Option to have the canvas be treated like a toroid by wrapping around.
	wrap bool

	// Maintain the field on which the automata live
	width int
	field [][]int
}

type tickMsg time.Time

var wrapFlag = flag.Bool("wrap", false, "Wrap the grid at the edges, treating it like a toroid")

// wrapPos wraps a cell position that would otherwise be outside of a rectangular grid.
func (m model) wrapPos(pos int) int {
	return int(math.Abs(float64(pos))) % len(m.field)
}

// nextGeneration evolves the field of automata one generation based on the rules of Conway's Game of Life.
func nextGeneration(m model) model {
	// Create a new field based on the existing one.
	next := make([][]int, len(m.field))
	for i := 0; i < len(m.field); i++ {
		next[i] = make([]int, len(m.field[0]))
	}

	// Loop over rows...
	for y, _ := range m.field {

		// Loop over columns...
		for x, _ := range m.field[y] {
			neighborCount := 0

			// Count the adjacent living cells on the row above.
			if y-1 >= 0 {
				if x-1 >= 0 {
					neighborCount += m.field[y-1][x-1]
				}
				neighborCount += m.field[y-1][x]
				if x+1 < len(m.field[y]) {
					neighborCount += m.field[y-1][x+1]
				}
			}

			// Count the adjacent cells to either side.
			if x-1 >= 0 {
				neighborCount += m.field[y][x-1]
			}
			if x+1 < len(m.field[y]) {
				neighborCount += m.field[y][x+1]
			}

			// Count the adjacent cells on the row below.
			if y+1 < len(m.field) {
				if x-1 >= 0 {
					neighborCount += m.field[y+1][x-1]
				}
				neighborCount += m.field[y+1][x]
				if x+1 < len(m.field[y]) {
					neighborCount += m.field[y+1][x+1]
				}
			}

			// Evolve the current cell by the following rules:
			//
			// 1. A dead cell becomes live if it's surrounded by exactly three living cells to represent breeding.
			// 2. A living cell dies of loneliness if it has 0 or 1 neighbors.
			// 3. A living cell dies of overcrowding if it has more than 3 neighbors.
			// 4. A living cell stays alive if it has 2 or 3 neighbors.
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
	m.field = next
	return m
}

// generateField generates a random field of automata, where each cell has a 1 in 5 chance of being alive.
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
		switch msg.String() {

		// Quit on Escape or Ctrl+C
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		// Regenerate the field on Ctrl+R
		case "ctrl+r":
			m = generateField(m)
			return m, nil
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			if m.field[msg.Y][msg.X] == 1 {
				m.field[msg.Y][msg.X] = 0
			} else {
				m.field[msg.Y][msg.X] = 1
			}
		}
		return m, nil

	// Window size messages — we receive one initially, and then again on every resize
	case tea.WindowSizeMsg:

		// Reset the field to the correct size
		m.field = make([][]int, msg.Height)
		for i := 0; i < msg.Height; i++ {
			m.field[i] = make([]int, msg.Width)
		}

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
	for _, row := range m.field {

		// Loop over collumns...
		for _, col := range row {

			// Set the cell contents
			if col == 1 {
				frame += "•"
			} else {
				frame += " "
			}
		}

		// Newline at the end of every row.
		frame += "\n"
	}
	return frame
}

func main() {
	flag.Parse()
	p := tea.NewProgram(model{
		wrap: *wrapFlag,
	}, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
