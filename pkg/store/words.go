package store

import (
	"container/heap"
	"fmt"
	"time"
)

const (
	minScore = 0
	maxScore = 5
)

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
