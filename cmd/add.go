package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/egregors/karten/store"
)

const (
	addMode = iota
	saveMode
	manualMode
)

type modelAdd struct {
	mode int

	words store.Words
	save  func(store.Words) error

	textInput textinput.Model

	w *store.Word
}

func (m modelAdd) Init() tea.Cmd {
	return textinput.Blink
}

func (m modelAdd) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyCtrlC:
			err := m.save(m.words)
			// todo: err
			if err != nil {
				fmt.Println("ERR:", err.Error())
			}
			return m, tea.Quit

		case tea.KeyEsc:
			if m.mode == manualMode {
				m.textInput.Reset()
				m.textInput.Placeholder = "Neues Wort..."
				m.mode = addMode
				return m, nil
			}

		case tea.KeyEnter:
			switch m.mode {
			case addMode:
				m.w = store.NewWord(m.textInput.Value())

				// can't get VF card
				if m.w.Translation == "" {
					m.textInput.Reset()
					m.textInput.Placeholder = "die Übersetzung..."
					m.mode = manualMode
					return m, nil
				}
				m.mode = saveMode
				return m, nil

			case manualMode:
				// todo: add a way to cancel word adding
				m.w.Translation = m.textInput.Value()
				m.textInput.Reset()
				m.textInput.Placeholder = "Neues Wort..."
				m.mode = saveMode
				return m, nil

			case saveMode:
				// todo: add a way to cancel translation proposal
				m.words = append(m.words, m.w)
				m.textInput.Reset()
				m.mode = addMode
				return m, nil
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m modelAdd) View() string {
	return strings.Join(
		[]string{
			m.titleWidget(),
			m.inputWidget(),
			m.helpWidget(),
		},
		"\n",
	)
}

func (m modelAdd) titleWidget() string {
	return "Welches Wort möchtest du lernen?\n"
}

func (m modelAdd) inputWidget() string {
	// todo: make all three parts looks like a Card.
	// 	With center positioning.
	var s string
	switch m.mode {
	case addMode:
		s = m.textInput.View()
	case manualMode:
		s = m.w.Origin + m.textInput.View()
	case saveMode:
		origin := m.w.StyledString(greenStyle, blueStyle)
		tr := m.w.Translation
		if tr == "" {
			tr = m.textInput.View()
		}
		s = origin + "\n" + wordStyle(tr)
	}
	return s
}

func (m modelAdd) helpWidget() string {
	msg := "\nctrl+c: quit • enter: "
	switch m.mode {
	case addMode:
		msg += "get translation"
	case manualMode:
		msg += "set translation"
	case saveMode:
		msg += "save"
	}
	return helpStyle(msg)
}

// RunAddModeCLI runs CLI in add-mode // todo: don't like func name
func RunAddModeCLI(path string) error {
	ws := store.Words{}
	_, err := ws.LoadFromCSV(path)
	if err != nil {
		return err
	}

	// todo: add smth like GetWordsWithSave callback maybe
	save := func(ws store.Words) error {
		_, err := ws.SaveToCSV(path)
		return err
	}

	m := modelAdd{
		words:     ws,
		save:      save,
		textInput: makeTextInput(),
	}

	if err := tea.NewProgram(m).Start(); err != nil {
		return err
	}

	return nil
}

func makeTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Neues Wort..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return ti
}
