package cmd

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/egregors/karten/store"
	"github.com/muesli/termenv"
)

var (
	color       = termenv.EnvColorProfile().Color
	helpStyle   = termenv.Style{}.Foreground(color("241")).Styled
	wordStyle   = termenv.Style{}.Foreground(color("150")).Styled
	finishStyle = termenv.Style{}.Foreground(color("10")).Styled

	greenStyle = termenv.Style{}.Foreground(color("46")).Styled
	blueStyle  = termenv.Style{}.Foreground(color("69")).Styled
)

const (
	scoreMarkOn  = "⭐️"
	scoreMarkOff = "✖️"
)

type model struct {
	// todo: rename words to sessionWords
	sessionWords store.Words
	currWord     *store.Word

	forgotten, memorized []*store.Word

	save func() error
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.currWord == nil {
		return exit(m)
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return exit(m)

		case "up":
			// don't remember
			m.currWord.DecScore()
			m.forgotten = append(m.forgotten, m.currWord)
			m.currWord = m.sessionWords.Next()

		case "down":
			// remember
			m.currWord.IncScore()
			m.memorized = append(m.memorized, m.currWord)
			m.currWord = m.sessionWords.Next()
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.currWord == nil {
		return strings.Join([]string{
			m.titleWidget(),
			m.forgottenWidget(),
			finishStyle("    🎉🎉🎉 Gut gemacht!\n"),
			m.memorizedWidget(),
			helpStyle("\n\n  any key: save and quit"),
		}, "\n")
	}

	frame := []string{
		m.titleWidget(),
		m.forgottenWidget(),
		m.wordWidget() + " | " + m.verbFormenWidget(),
		m.memorizedWidget(),
		m.helpWidget(),
	}

	return strings.Join(frame, "\n")
}

func exit(m model) (tea.Model, tea.Cmd) {
	// fixme: get path from outside
	err := m.save()
	if err != nil {
		fmt.Printf("ERR: %v", err.Error())
	}
	return m, tea.Quit
}

func (m model) getScoreStars() string {
	var score string
	for i := 0; i < 5; i++ {
		if i < m.currWord.Score {
			score += scoreMarkOn
		} else {
			score += scoreMarkOff
		}
	}
	return score
}

func (m model) verbFormenWidget() string {
	if m.currWord.Card == nil {
		return ""
	}

	res := ""
	for _, w := range m.currWord.Card.Forms {
		sub := w.Val
		switch w.Color {
		case store.Green:
			res += greenStyle(sub)
		case store.Blue:
			res += blueStyle(sub)
		default:
			res += wordStyle(sub)
		}
	}

	return res + "\n"
}

func (m model) helpWidget() string {
	return helpStyle("\n  up: ich habe das vergessen • down: ich erinnere mich • q | ctrl+c | esc: exit\n")
}

func (m model) wordWidget() string {
	score := m.getScoreStars()
	return fmt.Sprintf("    %s  %s", score, wordStyle(m.currWord.Origin))
}

func (m model) forgottenWidget() string {
	ws := make([]string, len(m.forgotten))
	for i, w := range m.forgotten {
		ws[i] = fmt.Sprintf("    %s - %s", w.Origin, w.Translation)
	}
	return strings.Join(ws, "\n") + "\n"
}

func (m model) memorizedWidget() string {
	start := 250
	s := func(c int, s string) string {
		clr := strconv.Itoa(c)
		return termenv.Style{}.Foreground(color(clr)).Styled(s)
	}
	var ws []string
	// todo: extract 5 to consts
	for i := len(m.memorized) - 1; i >= 0 && len(m.memorized)-i <= 5; i-- {
		ws = append(ws, s(start, fmt.Sprintf("    %s - %s", m.memorized[i].Origin, m.memorized[i].Translation)))
		start -= 3
	}
	return strings.Join(ws, "\n")
}

func (m model) titleWidget() string {
	return "  Kennst du dieses Wort?\n"
}

// RunCLI starts CLI client with words in `path` and session size `size`.
// Session size is how many words will be taken for learning.
func RunCLI(path string) error {
	allWords, sessionWords := store.Words{}, store.Words{}

	_, err := allWords.LoadFromCSV(path)
	if err != nil {
		return err
	}

	for _, w := range allWords {
		if w.Score < 5 {
			sessionWords = append(sessionWords, w)
		}
	}

	// todo: add smth like GetWordsWithSave callback maybe
	save := func() error {
		_, err := allWords.SaveToCSV(path)
		return err
	}

	m := model{
		sessionWords: sessionWords,
		currWord:     sessionWords.Next(),

		save: save,
	}

	if err := tea.NewProgram(m).Start(); err != nil {
		return err
	}
	return nil
}
