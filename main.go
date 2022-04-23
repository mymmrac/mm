package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mymmrac/mm/utils"
)

const (
	historyNone     = -1
	historyDisabled = -2
)

type Debugger struct {
	text    string
	enabled bool
}

func (d *Debugger) SetEnabled(enabled bool) {
	d.enabled = enabled
}

func (d *Debugger) Enabled() bool {
	return d.enabled
}

func (d *Debugger) Debug(args ...any) {
	d.text += fmt.Sprint(args...) + "\n"
}

func (d *Debugger) Clean() {
	d.text = ""
}

func (d *Debugger) Text() string {
	return d.text
}

type model struct {
	input textinput.Model

	liveResult   string
	liveError    bool
	liveErrorLoc Location

	selectedExpr int
	expressions  []string
	results      []string

	executor  *Executor
	exprError *ExprError

	debugger *Debugger

	width, height int
}

func newModel() *model {
	input := textinput.New()
	input.Placeholder = "..."
	input.Prompt = "> "
	input.Focus()

	debugger := &Debugger{}
	debugger.SetEnabled(false)

	return &model{
		input:        input,
		expressions:  make([]string, 0),
		selectedExpr: historyNone,
		executor:     NewExecutor(debugger),
		debugger:     debugger,
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyUpdate := false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type != tea.KeyLeft && msg.Type != tea.KeyRight {
			m.exprError = nil
			keyUpdate = true
		}

		switch {
		case key.Matches(msg, keys.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, keys.Quit):
			if m.input.Value() == "" {
				return m, tea.Quit
			}

			m.input.SetValue("")
			m.selectedExpr = historyNone
		case key.Matches(msg, keys.Execute):
			expr := m.input.Value()
			if strings.TrimSpace(expr) == "" {
				break
			}

			result, err := m.executor.Execute(expr)
			if err != nil {
				m.exprError = err
				m.input.SetCursor(err.loc.end)
				m.selectedExpr = historyDisabled
				break
			}

			m.expressions = append(m.expressions, expr)
			m.results = append(m.results, result)

			m.input.SetValue("")
			m.selectedExpr = historyNone
		case key.Matches(msg, keys.PrevExpr):
			if len(m.expressions) == 0 {
				break
			}

			if m.selectedExpr < 0 {
				m.selectedExpr = len(m.expressions) - 1
			} else if m.selectedExpr > 0 {
				if m.expressions[m.selectedExpr] != m.input.Value() {
					m.selectedExpr = historyDisabled
					break
				}
				m.selectedExpr--
			} else {
				break
			}

			m.input.SetValue(m.expressions[m.selectedExpr])
			m.input.CursorEnd()
		case key.Matches(msg, keys.NextExpr):
			if m.selectedExpr >= 0 {
				if m.expressions[m.selectedExpr] != m.input.Value() {
					m.selectedExpr = historyDisabled
					break
				}
				m.selectedExpr++
			} else {
				break
			}

			if m.selectedExpr == len(m.expressions) {
				m.selectedExpr = historyNone
				m.input.SetValue("")
				break
			}

			m.input.SetValue(m.expressions[m.selectedExpr])
			m.input.CursorEnd()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.selectedExpr == historyDisabled {
		keys.PrevExpr.SetEnabled(false)
		keys.NextExpr.SetEnabled(false)
	}

	if m.input.Value() == "" {
		keys.PrevExpr.SetEnabled(true)
		keys.NextExpr.SetEnabled(true)
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)

	if keyUpdate {
		liveResult, err := m.executor.Execute(m.input.Value())
		if err != nil {
			m.liveResult = ""
			m.liveError = true
			m.liveErrorLoc = err.loc
		} else {
			m.liveResult = liveResult
			m.liveError = false
		}
	}

	return m, inputCmd
}

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m *model) View() string {
	s := strings.Builder{}

	s.WriteString("\n")

	for i, expr := range m.expressions {
		s.WriteString(utils.Wrap("> "+expr+"\n", m.width))
		s.WriteString(utils.Wrap("=> "+m.results[i]+"\n\n", m.width))
	}

	m.input.Width = m.width - len(m.input.Prompt) - 1
	s.WriteString(m.input.View())

	if m.liveError || m.exprError != nil {
		loc := m.liveErrorLoc
		if m.exprError != nil {
			loc = m.exprError.loc
		}

		s.WriteString(fmt.Sprintf(
			"\n%s%s\n",
			strings.Repeat(" ", loc.start+len(m.input.Prompt)),
			errorStyle.Render(strings.Repeat("^", loc.Size())),
		))
	} else if m.liveResult != "" {
		s.WriteString(utils.Wrap(mutedStyle.Render("\n=> "+m.liveResult+"\n"), m.width))
	}

	if m.exprError != nil {
		s.WriteString(utils.Wrap("Error: "+m.exprError.text, m.width))
	}

	if m.debugger.Enabled() {
		s.WriteString(strings.Repeat("\n", 3) + utils.Wrap(m.debugger.Text(), m.width))
	}

	return s.String()
}

func main() {
	if err := tea.NewProgram(newModel()).Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "FATAL: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Bye!")
}
