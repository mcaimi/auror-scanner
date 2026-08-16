package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types sent from the analysis goroutine.
type StdoutMsg string
type StderrMsg string
type DoneMsg struct{ Err error }

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("252")).
			PaddingLeft(1)

	separatorColor = lipgloss.Color("240")
)

type model struct {
	pkgfile   string
	baseURL   string
	modelName string

	stdoutVP  viewport.Model
	stderrVP  viewport.Model
	stdoutBuf []string
	stderrBuf []string

	width  int
	height int
	ready  bool
	done   bool
	err    error
}

func newModel(pkgfile, baseURL, modelName string) model {
	return model{
		pkgfile:   pkgfile,
		baseURL:   baseURL,
		modelName: modelName,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeViewports()
		m.ready = true
		return m, nil

	case StdoutMsg:
		m.stdoutBuf = append(m.stdoutBuf, string(msg))
		m.stdoutVP.SetContent(strings.Join(m.stdoutBuf, "\n"))
		m.stdoutVP.GotoBottom()
		return m, nil

	case StderrMsg:
		m.stderrBuf = append(m.stderrBuf, string(msg))
		m.stderrVP.SetContent(strings.Join(m.stderrBuf, "\n"))
		m.stderrVP.GotoBottom()
		return m, nil

	case DoneMsg:
		m.done = true
		m.err = msg.Err
		return m, nil
	}

	// Forward remaining messages to viewports (handles scroll keys).
	var cmd tea.Cmd
	m.stdoutVP, cmd = m.stdoutVP.Update(msg)
	cmds = append(cmds, cmd)
	m.stderrVP, cmd = m.stderrVP.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// resizeViewports recalculates pane dimensions and rebuilds both viewports.
// Top pane gets 60%, bottom pane gets 40%, with one separator row between them.
func (m model) resizeViewports() model {
	centerH := m.height - 2
	if centerH < 1 {
		centerH = 1
	}
	avail := centerH - 1 // one row reserved for the border separator
	topH := int(float64(avail) * 0.6)
	bottomH := avail - topH

	m.stdoutVP = viewport.New(m.width, topH)
	m.stdoutVP.SetContent(strings.Join(m.stdoutBuf, "\n"))
	m.stdoutVP.GotoBottom()

	m.stderrVP = viewport.New(m.width, bottomH)
	m.stderrVP.SetContent(strings.Join(m.stderrBuf, "\n"))
	m.stderrVP.GotoBottom()

	return m
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// ── top bar ─────────────────────────────────────────────────────────────
	header := headerStyle.Width(m.width).Render("AUROR - " + m.pkgfile)

	// ── bottom bar ───────────────────────────────────────────────────────────
	statusLine := m.baseURL + " - " + m.modelName
	switch {
	case m.err != nil:
		statusLine += "  [ERROR: " + m.err.Error() + "]"
	case m.done:
		statusLine += "  [DONE — press q to quit]"
	default:
		statusLine += "  [RUNNING]"
	}
	footer := footerStyle.Width(m.width).Render(statusLine)

	// ── center: top (stdout) / separator / bottom (stderr) ──────────────────
	centerH := m.height - 2
	avail := centerH - 1 // one row reserved for the border separator
	topH := int(float64(avail) * 0.6)
	bottomH := avail - topH

	topPane := lipgloss.NewStyle().
		Width(m.width).
		Height(topH).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(separatorColor).
		Render(m.stdoutVP.View())

	bottomPane := lipgloss.NewStyle().
		Width(m.width).
		Height(bottomH).
		Render(m.stderrVP.View())

	center := lipgloss.JoinVertical(lipgloss.Left, topPane, bottomPane)

	return lipgloss.JoinVertical(lipgloss.Left, header, center, footer)
}
