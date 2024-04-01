package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO:
// * https://github.com/lrstanley/bubblezone for mouse support on tabs and textareas
// * Cursor blink

type tickMsg time.Time

// tick updates the model every 1/10 second.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func tabBorderWithTop(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.TopLeft = left
	border.Top = middle
	border.TopRight = right
	return border
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	stBlue  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111"))
	stPink  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	stWhite = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("253"))
	stBlink = stWhite.Copy().Blink(true)
	stTitle = stWhite.Render("──┤ ") +
		stBlue.Render("St") +
		stPink.Render("im") +
		stWhite.Render("mta") +
		stPink.Render("us") +
		stBlue.Render("ch") +
		stWhite.Render(" ├")

	stTabBorders   = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	stInputBorder  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	stActive       = lipgloss.NewStyle().Foreground(lipgloss.Color("253"))
	stInactive     = lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	stDisconnected = lipgloss.NewStyle().Foreground(lipgloss.Color("95"))

	inactiveTabBorder            = tabBorderWithTop("┬", "─", "┬")
	activeTabBorder              = tabBorderWithTop("┐", " ", "┌")
	inactiveTabStyle             = stInactive.Copy().Border(inactiveTabBorder, true).BorderForeground(stTabBorders.GetForeground()).Padding(0, 1)
	activeTabStyle               = stActive.Copy().Border(activeTabBorder, true).BorderForeground(stTabBorders.GetForeground()).Padding(0, 1)
	inactiveDisconnectedTabStyle = stDisconnected.Copy().Border(inactiveTabBorder, true).BorderForeground(stTabBorders.GetForeground()).Padding(0, 1)
	inputBorder                  = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, false, true).BorderForeground(stInputBorder.GetForeground())
)

type view struct {
	title                           string
	current, activity, disconnected bool
	input                           textarea.Model
	output                          viewport.Model
}

type stui struct {
	title, tabs, chars string
	views              []view
	worldIndex         int
	ready              bool
}

func (st stui) updateTitle() stui {
	t := time.Now()
	info := stWhite.Render(fmt.Sprintf("┤ %3.f%% ├───┤ ", st.views[st.worldIndex].output.ScrollPercent()*100)) +
		stBlue.Render(fmt.Sprintf("%2d", t.Hour())) + stBlink.Render(":") + stPink.Render(fmt.Sprintf("%02d", t.Minute())) +
		stWhite.Render(" ├──")
	st.title = lipgloss.JoinHorizontal(lipgloss.Center,
		stTitle,
		stWhite.Render(strings.Repeat("─", max(3, st.views[st.worldIndex].output.Width-lipgloss.Width(stTitle)-lipgloss.Width(info)))),
		info)
	return st
}

func (st stui) updateTabs() stui {
	var renderedTabs []string

	for i, t := range st.views {
		style := inactiveTabStyle.Copy()
		if t.current {
			style = activeTabStyle.Copy()
			if t.disconnected {
				style = style.Foreground(stDisconnected.GetForeground())
			}
		} else if t.disconnected {
			style = inactiveDisconnectedTabStyle.Copy()
		}
		if t.activity {
			style = style.Underline(true).Bold(true)
		}
		border, _, _, _, _ := style.GetBorder()
		if i == 0 && t.current {
			border.TopLeft = "│"
		} else if i == 0 && !t.current {
			border.TopLeft = "├"
		} else if i == len(st.views)-1 && t.current {
			border.TopRight = "│"
		} else if i == len(st.views)-1 && !t.current {
			border.TopRight = "┤"
		}
		renderedTabs = append(renderedTabs, style.Render(t.title))
	}
	st.tabs = lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
	st.tabs = lipgloss.JoinHorizontal(
		lipgloss.Top, stTabBorders.Render("───"),
		st.tabs,
		stTabBorders.Render(strings.Repeat("─", max(3, st.views[st.worldIndex].output.Width-lipgloss.Width(st.tabs)))))
	return st
}

func (st stui) confirmQuit() tea.Cmd {
	return tea.Quit
}

func (st stui) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	st.views[st.worldIndex].input, cmd = st.views[st.worldIndex].input.Update(msg)
	return cmd
}

func (st stui) Init() tea.Cmd {
	return tick()
}

func (st stui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	switch msg := msg.(type) {
	case tickMsg:
		st = st.updateTitle()
		return st, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return st, st.confirmQuit()
		case tea.KeyCtrlPgDown:
			st.views[st.worldIndex].current = false
			st.worldIndex++
			if st.worldIndex >= len(st.views) {
				st.worldIndex = 0
			}
			st.views[st.worldIndex].current = true
			cmds = append(cmds, viewport.Sync(st.views[st.worldIndex].output))
		case tea.KeyPgDown:
			lines := st.views[st.worldIndex].output.ViewDown()
			st.updateTitle()
			return st, viewport.ViewDown(st.views[st.worldIndex].output, lines)
		case tea.KeyCtrlPgUp:
			st.views[st.worldIndex].current = false
			st.worldIndex--
			if st.worldIndex < 0 {
				st.worldIndex = len(st.views) - 1
			}
			st.views[st.worldIndex].current = true
			cmds = append(cmds, viewport.Sync(st.views[st.worldIndex].output))
		case tea.KeyPgUp:
			lines := st.views[st.worldIndex].output.ViewUp()
			st.updateTitle()
			return st, viewport.ViewUp(st.views[st.worldIndex].output, lines)
		}
	case tea.WindowSizeMsg:
		if !st.ready {
			for i, _ := range st.views {
				v := viewport.New(msg.Width, msg.Height-11)
				v.YPosition = 3
				v.SetContent(st.views[i].title)
				v.HighPerformanceRendering = true
				t := st.views[i].input
				t.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "252", Dark: "235"})
				t.MaxWidth = msg.Width - 2
				t.SetWidth(msg.Width - 2)
				t.SetHeight(3)
				t.Placeholder = st.views[i].title
				t.Prompt = ""
				if i == st.worldIndex {
					t.Focus()
				} else {
					t.Blur()
				}
				st.views[i].output = v
				st.views[i].input = t
			}
			st.ready = true
		} else {
			for _, v := range st.views {
				v.output.Width = msg.Width
				v.output.Height = msg.Height - 11
				v.input.MaxWidth = msg.Width - 2
				v.input.SetWidth(msg.Width - 2)
			}
		}
		cmds = append(cmds, viewport.Sync(st.views[st.worldIndex].output))
	}
	st = st.updateTabs()
	st = st.updateTitle()
	st.views[st.worldIndex].input.Focus()
	st.views[st.worldIndex].input, cmd = st.views[st.worldIndex].input.Update(msg)
	cmds = append(cmds, cmd)
	st.views[st.worldIndex].output, cmd = st.views[st.worldIndex].output.Update(msg)
	cmds = append(cmds, cmd)
	charCount := st.views[st.worldIndex].input.Length()
	s := "s"
	if charCount == 1 {
		s = ""
	}
	st.chars = fmt.Sprintf("%d char%s", charCount, s)
	st.chars = stInputBorder.Render("└"+strings.Repeat("─", st.views[st.worldIndex].output.Width-lipgloss.Width(st.chars)-8)+"┤ ") + st.chars + stInputBorder.Render(" ├──┘")
	return st, tea.Batch(cmds...)
}

func (st stui) View() string {
	return fmt.Sprintf("%s\n\n%s\n%s\n%s\n\n%s",
		st.title,
		st.views[st.worldIndex].output.View(),
		inputBorder.Render(st.views[st.worldIndex].input.View()),
		st.chars,
		st.tabs)
}

func initializeModel() stui {
	return stui{
		views: []view{
			view{
				title:   "Current",
				current: true,
				input:   textarea.New(),
				output:  viewport.New(0, 0),
			},
			view{
				title:  "Inactive",
				input:  textarea.New(),
				output: viewport.New(0, 0),
			},
			view{
				title:    "Inactive (activity)",
				activity: true,
				input:    textarea.New(),
				output:   viewport.New(0, 0),
			},
			view{
				title:        "Disconnected",
				disconnected: true,
				input:        textarea.New(),
				output:       viewport.New(0, 0),
			},
			view{
				title:        "Disconnected (activity)",
				disconnected: true,
				activity:     true,
				input:        textarea.New(),
				output:       viewport.New(0, 0),
			},
		},
	}
}

func main() {
	p := tea.NewProgram(initializeModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
