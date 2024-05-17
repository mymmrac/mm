package repl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mymmrac/mm/debugger"
	"github.com/mymmrac/mm/executor/v2"
	"github.com/mymmrac/mm/utils"
)

const (
	historyNone     = -1
	historyDisabled = -2
)

type Model struct {
	input textinput.Model

	liveResult   string
	liveError    bool
	liveErrorLoc executor.Location

	selectedExpr int
	expressions  []string
	results      []string

	executor  *executor.Executor
	precision int32
	error     error
	exprError *executor.ExprError

	debugger *debugger.Debugger

	width, height int
}

func NewModel(debugger *debugger.Debugger, precision int32) *Model {
	input := textinput.New()
	input.Placeholder = "..."
	input.Prompt = "> "
	input.Focus()

	return &Model{
		input:        input,
		expressions:  make([]string, 0),
		selectedExpr: historyNone,
		executor:     executor.NewExecutor(debugger),
		precision:    precision,
		debugger:     debugger,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(rawMsg tea.Msg) (tea.Model, tea.Cmd) {
	keyUpdate := false

	switch msg := rawMsg.(type) {
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

			result, err := m.executor.Execute(expr, m.precision)
			if err != nil {
				m.error = err
				m.selectedExpr = historyDisabled

				if errors.As(err, &m.exprError) {
					m.input.SetCursor(m.exprError.Loc.End)
				} else {
					m.exprError = nil
				}

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
			if key.Matches(msg, keys.UseResult) && len(m.results) != 0 && m.input.Value() == "" {
				lastResult := m.results[len(m.results)-1]
				m.input.SetValue(lastResult)

				break
			}

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
	m.input, inputCmd = m.input.Update(rawMsg)

	if keyUpdate {
		liveResult, err := m.executor.Execute(m.input.Value(), m.precision)
		if err != nil {
			if errors.As(err, &m.exprError) {
				m.liveResult = ""
				m.liveError = true
				m.liveErrorLoc = m.exprError.Loc
				m.selectedExpr = historyNone
			} else {
				m.error = err
				m.exprError = nil
				m.selectedExpr = historyDisabled
			}
		} else {
			m.liveResult = liveResult
			m.liveError = false
		}
	}

	if _, ok := rawMsg.(cursor.BlinkMsg); !ok {
		m.debugger.Debug("Message", fmt.Sprintf(" %#v", rawMsg))
	}

	return m, inputCmd
}

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m *Model) View() string {
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
			loc = m.exprError.Loc
		}

		s.WriteString(fmt.Sprintf(
			"\n%s%s\n",
			strings.Repeat(" ", loc.Start+len(m.input.Prompt)),
			errorStyle.Render(strings.Repeat("^", loc.Size())),
		))
	} else if m.liveResult != "" {
		s.WriteString(utils.Wrap(mutedStyle.Render("\n=> "+m.liveResult+"\n"), m.width))
	}

	if m.exprError != nil {
		s.WriteString(utils.Wrap("Error: "+m.exprError.Message, m.width))
	} else if m.error != nil {
		s.WriteString(utils.Wrap("Error: "+m.error.Error(), m.width))
	}

	if m.debugger.Enabled() {
		s.WriteString(strings.Repeat("\n", 3) + utils.Wrap(m.debugger.String(), m.width))
	}

	return s.String()
}
