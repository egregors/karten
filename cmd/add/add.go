package add

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/egregors/karten/pkg/store"
	"github.com/egregors/karten/pkg/widgets"
)

// WordAdder is persistent store for new words
type WordAdder interface {
	// AddWord adds a new store.Word into store
	AddWord(w *store.Word) error
}

// MetaProvider is remote meta provider to get some Meta data for store.Word
type MetaProvider interface {
	// GetMeta perform request to MetaProvider and get meta as a string
	// if it'S possible
	GetMeta(w *store.Word) error
}

// Srv is service to add new words
type Srv struct {
	Store    WordAdder
	Provider MetaProvider

	UI *tea.Program

	dbg bool
}

// NewSrv creates a new service to adding new words
func NewSrv(s WordAdder, p MetaProvider, dbg bool) *Srv {
	srv := &Srv{
		Store:    s,
		Provider: p,
		dbg:      dbg,
	}

	srv.UI = tea.NewProgram(addModel{
		Mode:      addMode,
		TextInput: makeTextInput(),
		S:         srv,
	})

	return srv
}

// Run starts CLI interface
func (srv *Srv) Run() error {
	return srv.UI.Start()
}

const (
	// UI modes
	addMode    = iota // add word (active origin input)
	saveMode          // save word (got translation from MetaProvider, or from user in manual Mode)
	manualMode        // set translation by own hands
)

type addModel struct {
	S *Srv

	TextInput textinput.Model

	Mode        int
	CurrentWord *store.Word

	CurrErr error
}

func (m addModel) GetCurrErr() string {
	if m.CurrErr != nil {
		return m.CurrErr.Error()
	}
	return ""
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if m.Mode == manualMode {
				m.Mode = addMode
				m.updateTextInput()
				return m, cmd
			}

		case tea.KeyEnter:
			switch m.Mode {
			case addMode:
				if m.TextInput.Value() == "" {
					break
				}

				m.CurrentWord = store.NewWord(m.TextInput.Value())
				err := m.S.Provider.GetMeta(m.CurrentWord)

				if err != nil {
					m.CurrErr = err
					m.Mode = manualMode
					m.updateTextInput()
					return m, cmd
				}

				m.Mode = saveMode
				m.updateTextInput()
				return m, cmd

			case manualMode:
				m.CurrentWord.Translation = m.TextInput.Value()
				m.Mode = saveMode
				m.updateTextInput()
				return m, cmd

			case saveMode:
				err := m.S.Store.AddWord(m.CurrentWord)
				if err != nil {
					// todo: handle error in proper way
					m.CurrErr = err
					return m, cmd
				}
				m.Mode = addMode
				m.updateTextInput()
				return m, cmd
			}
		}
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m addModel) View() string {
	frame := []string{
		m.titleWidget(),
		m.inputWidget(),
		m.helpWidget(),
	}

	if m.S.dbg {
		frame = append(frame, widgets.DebugWidget(m))
	}

	return strings.Join(frame, "\n")
}

func (m addModel) titleWidget() string {
	return ">>> Karten 🃏\n"
}

func (m addModel) inputWidget() string {
	// todo: make all three parts looks like a Card.
	// 	With center positioning.
	var s string
	switch m.Mode {
	case addMode:
		s = m.TextInput.View()
	case manualMode:
		s = m.CurrentWord.Origin + " –" + m.TextInput.View()
	case saveMode:
		origin := m.CurrentWord.Origin
		tr := m.CurrentWord.Translation
		if tr == "" {
			tr = m.TextInput.View()
		}
		s = origin + " – " + tr

		if m.CurrentWord.Meta != "" {
			s += "\n\n" + m.CurrentWord.Meta
		}
	}
	return s
}

func (m addModel) helpWidget() string {
	msg := "\nctrl+c: quit • enter: "
	switch m.Mode {
	case addMode:
		msg += "get translation"
	case manualMode:
		msg += "set translation • esc: cancel  f"
	case saveMode:
		msg += "save"
	}
	return msg
}

func (m *addModel) updateTextInput() {
	switch m.Mode {
	case manualMode:
		m.TextInput.Reset()
		m.TextInput.Placeholder = "Translation..."
	case addMode, saveMode:
		m.TextInput.Reset()
		m.TextInput.Placeholder = "New word..."
	}
}

func makeTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "New word..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return ti
}
