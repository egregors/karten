package widgets

import (
	"fmt"
	"reflect"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	color           = termenv.EnvColorProfile().Color
	mainStyle       = termenv.Style{}.Foreground(color("11")).Styled
	modelFieldStyle = termenv.Style{}.Foreground(color("39")).Styled
	modelValStyle   = termenv.Style{}.Foreground(color("87")).Styled
	errStyle        = termenv.Style{}.Foreground(color("1")).Styled
	nonErrStyle     = termenv.Style{}.Foreground(color("83")).Styled
)

// DebugModel is model supports debugging widget
type DebugModel interface {
	tea.Model
	GetCurrErr() string
}

// DebugWidget returns some useful for debugging information
func DebugWidget(m DebugModel) string {
	err := "ERR: "
	currErr := m.GetCurrErr()
	if currErr != "" {
		err += errStyle(currErr)
	} else {
		err += nonErrStyle("nil")
	}

	s := strings.Join([]string{
		mainStyle("= = = = = DEBUG = = = = ="),
		mainStyle("| ") + err,
		getModelFrame(m),
		mainStyle("= = = = = ----- = = = = ="),
	}, "\n")

	return s
}

func getModelFrame(m DebugModel) string {
	e := reflect.ValueOf(&m).Elem().Elem()
	n := e.NumField()
	res := make([]string, n)

	for i := 0; i < n; i++ {
		varName := e.Type().Field(i).Name
		varType := e.Type().Field(i).Type
		varValue := e.Field(i).Interface()

		res[i] = fmt.Sprintf(
			"%-40s %-20s %-30s",
			mainStyle("| ")+modelFieldStyle(varName),
			"["+varType.String()+"]",
			modelValStyle(fmt.Sprintf("%v", varValue)),
		)
	}

	return strings.Join(res, "\n")
}
