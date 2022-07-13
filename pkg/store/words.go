package store

import (
	"container/heap"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	minScore = 0
	maxScore = 5

	ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
)

var re = regexp.MustCompile(ansi)

// Word is word for learning with all required metadata.
type Word struct {
	Origin, Translation, Meta string
	LastSeenAt                time.Time
	Score                     int
}

// NewWord create a new Word instance, including try to get word metadata form
// VerbFormen.
func NewWord(raw string) *Word {
	return &Word{Origin: raw}
}

func (w Word) String() string {
	return w.Origin
}

// IncScore increases particular Word score
func (w *Word) IncScore() {
	if w.Score < maxScore {
		w.Score++
	}
	w.LastSeenAt = time.Now()
}

// DecScore decrease particular Word score
func (w *Word) DecScore() {
	if w.Score > minScore {
		w.Score--
	}
}

// HasMeta indicate if Word has Meta data
func (w *Word) HasMeta() bool {
	return w.Meta != ""
}

// GetMeta returns formatted Meta
func (w *Word) GetMeta() string {
	return strings.Join(strings.Split(w.Meta, "\n")[1:], "")
}

// GetMetaWithoutColors return pure string without ASCII colors
func (w *Word) GetMetaWithoutColors() string {
	return StripColors(w.Meta)
}

// StripColors cleans any color related unprintable symbols
func StripColors(s string) string {
	return re.ReplaceAllString(s, "")
}

// Words is a heap of Word's
type Words []*Word

func (ws *Words) String() string {
	// todo: cut ws if len is too big
	return fmt.Sprintf("%d words: %v...", ws.Len(), *ws)
}

// Len Less Swap Push Pop needs to implement heap interface
func (ws *Words) Len() int           { return len(*ws) }
func (ws *Words) Less(i, j int) bool { return (*ws)[i].Score < (*ws)[j].Score }
func (ws *Words) Swap(i, j int)      { (*ws)[i], (*ws)[j] = (*ws)[j], (*ws)[i] }

// Push adds Word into heap
func (ws *Words) Push(x any) { *ws = append(*ws, x.(*Word)) }

// Pop remove last word from heap and returns it
func (ws *Words) Pop() any {
	x := (*ws)[ws.Len()-1]
	*ws = (*ws)[:ws.Len()-1]
	return x
}

// Next pops and returns next word
func (ws *Words) Next() (w *Word) {
	if !ws.IsEmpty() {
		w = heap.Pop(ws).(*Word)
	}
	return
}

// IsEmpty returns false if heap is empty
func (ws *Words) IsEmpty() bool { return ws.Len() == 0 }
