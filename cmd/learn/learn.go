package learn

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/egregors/karten/pkg/store"
	"github.com/egregors/karten/pkg/widgets"
	"github.com/muesli/termenv"
)

const (
	sessionSize = 20
)

var (
	color       = termenv.EnvColorProfile().Color
	helpStyle   = termenv.Style{}.Foreground(color("241")).Styled
	wordStyle   = termenv.Style{}.Foreground(color("150")).Styled
	finishStyle = termenv.Style{}.Foreground(color("10")).Styled

	goodStyle = termenv.Style{}.Foreground(color("46")).Styled
	badStyle  = termenv.Style{}.Foreground(color("69")).Styled
)

const (
	scoreMarkOn  = "⭐️"
	scoreMarkOff = "✖️"
)

// WordStore is store with store.Word for learning
type WordStore interface {
	// GetWords should return n words in score decreasing order
	GetWords(n int) (*store.Words, error)
	// Save commit current store.Word in the store
	Save(w *store.Word) error
}

// Srv is service to learn words
type Srv struct {
	Store WordStore

	UI *tea.Program

	dbg bool
}

// NewSrv creates a new service to learning words
func NewSrv(s WordStore, dbg bool) (*Srv, error) {
	srv := &Srv{
		Store: s,
		dbg:   dbg,
	}

	ws, err := srv.Store.GetWords(sessionSize)
	if err != nil {
		return nil, err
	}

	srv.UI = tea.NewProgram(learnModel{
		S:         srv,
		Words:     ws,
		CurrWord:  ws.Next(),
		Forgotten: []*store.Word{},
		Memorized: []*store.Word{},
	})

	return srv, nil
}

// Run starts CLI interface
func (srv *Srv) Run() error {
	return srv.UI.Start()
}

type learnModel struct {
	S *Srv

	Words    *store.Words
	CurrWord *store.Word

	Forgotten, Memorized []*store.Word

	CurrErr error
}

func (m learnModel) GetCurrErr() string {
	if m.CurrErr != nil {
		return m.CurrErr.Error()
	}
	return ""
}

func (m learnModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m learnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.CurrWord == nil {
		return m, tea.Quit
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "up":
			// don't remember
			m.CurrWord.DecScore()
			m.CurrErr = m.S.Store.Save(m.CurrWord)
			m.Forgotten = append(m.Forgotten, m.CurrWord)
			m.CurrWord = m.Words.Next()

		case "down":
			// remember
			m.CurrWord.IncScore()
			m.CurrErr = m.S.Store.Save(m.CurrWord)
			m.Memorized = append(m.Memorized, m.CurrWord)
			m.CurrWord = m.Words.Next()
		}
	}

	return m, nil
}

func (m learnModel) View() string {
	frame := []string{
		m.titleWidget(),
		m.forgottenWidget(),
		m.wordWidget(),
		m.memorizedWidget(),
		m.helpWidget(),
	}

	if m.S.dbg {
		frame = append(frame, widgets.DebugWidget(m))
	}
	return strings.Join(frame, "\n")
}

func (m learnModel) getScoreStars() string {
	var score string
	for i := 0; i < 5; i++ {
		if i < m.CurrWord.Score {
			score += scoreMarkOn
		} else {
			score += scoreMarkOff
		}
	}
	return score
}

func (m learnModel) titleWidget() string {
	return ">>> Karten 🃏\n"
}

func (m learnModel) wordWidget() string {
	if m.CurrWord == nil {
		return fmt.Sprintf(
			"    %s  %s / %s\n",
			finishStyle("Nice!"),
			goodStyle(strconv.Itoa(len(m.Memorized))),
			badStyle(strconv.Itoa(len(m.Forgotten))))
	}

	return fmt.Sprintf("    %s  %s\n", m.getScoreStars(), wordStyle(m.CurrWord.Origin))
}

func (m learnModel) forgottenWidget() string {
	ws := make([]string, len(m.Forgotten))
	for i, w := range m.Forgotten {
		ws[i] = fmt.Sprintf("    %s - %s", w.Origin, w.Translation)
	}
	return strings.Join(ws, "\n") + "\n"
}

func (m learnModel) memorizedWidget() string {
	start := 250
	s := func(c int, s string) string {
		clr := strconv.Itoa(c)
		return termenv.Style{}.Foreground(color(clr)).Styled(s)
	}
	var ws []string
	// todo: extract 5 to consts
	for i := len(m.Memorized) - 1; i >= 0 && len(m.Memorized)-i <= 5; i-- {
		ws = append(ws, s(start, fmt.Sprintf("    %s - %s", m.Memorized[i].Origin, m.Memorized[i].Translation)))
		start -= 3
	}
	return strings.Join(ws, "\n")
}

func (m learnModel) helpWidget() string {
	return helpStyle("\n  up: I know it! • down: i don't remember :( • q | ctrl+c | esc: exit\n")
}
