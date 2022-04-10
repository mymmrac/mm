package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

const (
	historyNone     = -1
	historyDisabled = -2
)

type model struct {
	input textinput.Model

	selectedExpr int
	expressions  []string
	results      []string

	executor  *Executor
	exprError *ExprError

	width, height int
}

func newModel() *model {
	input := textinput.New()
	input.Placeholder = "Expression..."
	input.Prompt = "> "
	input.Focus()

	return &model{
		input:        input,
		expressions:  make([]string, 0),
		selectedExpr: historyNone,
		executor:     NewExecutor(),
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type != tea.KeyLeft && msg.Type != tea.KeyRight {
			m.exprError = nil
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
	return m, inputCmd
}

func (m *model) View() string {
	s := strings.Builder{}

	s.WriteString("\n")

	for i, expr := range m.expressions {
		s.WriteString(wordwrap.String("> "+expr+"\n", m.width))
		s.WriteString(wordwrap.String("=> "+m.results[i]+"\n\n", m.width))
	}

	m.input.Width = m.width - len(m.input.Prompt) - 1
	s.WriteString(m.input.View())

	if m.exprError != nil {
		err := m.exprError

		s.WriteString(fmt.Sprintf(
			"\n%s%s\n",
			strings.Repeat(" ", err.loc.start+len(m.input.Prompt)),
			strings.Repeat("^", err.loc.Size()),
		))
		s.WriteString(wordwrap.String("Syntax error: "+err.text, m.width))
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
