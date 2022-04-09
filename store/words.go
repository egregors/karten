package store

import (
	"container/heap"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	minScore = 0
	maxScore = 5

	verbFormenURL = "https://www.verbformen.com/?w="
)

const (
	csvOrigin int = iota
	csvTranslation
	csvLastSeenAt
	csvScore
	csvVerbFormenJSON
)

//nolint:revive // it's just colors
const (
	Default int = iota
	Green
	Blue
)

// Syllable is the part words with a specific style (color).
// I want it to look like on VerbFormen (with word parts highlighting).
type Syllable struct {
	Val   string `json:"val"`
	Color int    `json:"color"`
}

// VerbFormenCard is representation of word card from VerbFormen
type VerbFormenCard struct {
	Origin      []string    `json:"origin"`
	Translation []string    `json:"translation"`
	Forms       []*Syllable `json:"forms"`
}

// NewVerbFormenCardFromNode takes HTML node with `section` from verbformen.com and
// parses it into VerbFormenCard struct
// todo: too complex, need refactoring
func NewVerbFormenCardFromNode(node *html.Node) *VerbFormenCard {
	// origin
	ps := findNode(node, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "p" {
			for _, a := range n.Attr {
				return strings.Contains(a.Val, "vGrnd")
			}
		}
		return false
	})

	textNodes := findNodes(ps, func(n *html.Node) bool {
		return n.Type == html.TextNode
	})

	var o []string
	for _, n := range textNodes {
		if n.Data != "\n" {
			o = append(o, strings.TrimSpace(strings.Trim(n.Data, "\n")))
		}
	}

	// word forms
	ps = findNode(node, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "p" {
			for _, a := range n.Attr {
				return strings.Contains(a.Val, "vStm")
			}
		}
		return false
	})

	textNodes = findNodes(ps, func(n *html.Node) bool {
		return n.Type == html.TextNode
	})

	var ws []*Syllable
	// todo: switch
	for _, n := range textNodes {
		d := strings.ReplaceAll(n.Data, "\n", " ")
		if d == "" || d == " " {
			continue
		}

		// neutral
		if n.Parent.Data == "p" || n.Parent.Data == "b" {
			ws = append(ws, &Syllable{d, Default})
		}

		// green
		if n.Parent.Data == "i" {
			ws = append(ws, &Syllable{d, Green})
		}

		// blue
		if n.Parent.Data == "u" {
			ws = append(ws, &Syllable{d, Blue})
		}
	}

	// translations
	ts := findNode(node, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, a := range n.Attr {
				return a.Key == "lang" && a.Val == "en"
			}
		}
		return false
	})

	textNodes = findNodes(ts, func(n *html.Node) bool {
		return n.Type == html.TextNode
	})

	var data string
	for _, n := range textNodes {
		if n.Data != "\n" {
			data = n.Data
		}
	}

	var t []string //nolint:prealloc // slice size is unknown
	for _, w := range strings.Split(data, "\n") {
		t = append(t, strings.TrimSuffix(strings.TrimSpace(w), ","))
	}

	return &VerbFormenCard{Origin: o, Translation: t, Forms: ws}
}

// Word is word for learning with all required metadata.
type Word struct {
	Origin, Translation string
	LastSeenAt          time.Time
	Score               int
	Card                *VerbFormenCard
}

// NewWord create a new Word instance, including try to get word metadata form
// VerbFormen.
func NewWord(raw string) *Word {
	// todo:
	// 	Here should be something like CardProvider.
	// 	The idea is provide an interface and flexible Word struct
	// 	to got ability to got words metadata from different providers
	// 	(like verbformen.com here) and for different languages.
	w := &Word{Origin: raw}
	card, err := w.GetVerbFormenCard()
	if err != nil {
		fmt.Printf("can't get VerbVormen card: %v", err)
		return w
	}
	if card != nil {
		w.Translation = strings.Join(card.Translation, ", ")
		w.Origin = strings.Join(card.Origin, " ")
	}
	w.Card = card

	return w
}

// NewWordFromCSV creates Word instances from string representation from the store.
// Actually, this is just an unmarshalling simple row from CSV.
func NewWordFromCSV(row []string) (*Word, error) {
	// todo:
	// 	I don't like name. It's more about unmarshalling not about creating a new one

	t, err := time.Parse(time.RFC3339, row[csvLastSeenAt])
	if err != nil {
		t = time.Time{}
	}

	score, err := strconv.Atoi(row[csvScore])
	if err != nil {
		score = 0
	}

	card := VerbFormenCard{}
	if row[csvVerbFormenJSON] != "" {
		err = json.Unmarshal([]byte(row[csvVerbFormenJSON]), &card)
		if err != nil {
			return nil, err
		}
	}

	return &Word{
		Origin:      row[csvOrigin],
		Translation: row[csvTranslation],
		LastSeenAt:  t,
		Score:       score,
		Card:        &card,
	}, nil
}

func (w Word) String() string {
	return w.Origin
}

// StyledString returns styled Word representation if it's possible.
// This representation is quite specific for VerbFormenProvider.
func (w Word) StyledString(green, blue func(string) string) string {
	s := w.String() + "\n"
	if w.Card != nil {
		for _, w := range w.Card.Forms {
			sub := w.Val
			switch w.Color {
			case Green:
				s += green(sub)
			case Blue:
				s += blue(sub)
			default:
				s += sub
			}
		}
	}
	return s
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

// GetJSON performs serialization into JSON for `Word.Forms`
func (v VerbFormenCard) GetJSON() string {
	if len(v.Forms) == 0 || len(v.Translation) == 0 {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		// fixme: err
		fmt.Println(err)
	}
	return string(data)
}

// GetCSV perform serialization for Word into CSV row.
//	 Schema:
//	 	origin 				:: string
//		translation 		:: string
//		last_seen_at 		:: string[time.RFC3339]
//		score 				:: int
//		meta 	:: string[JSON]
func (w Word) GetCSV() []string {
	// hint: schema
	var jsonCard string
	if w.Card != nil {
		jsonCard = w.Card.GetJSON()
	}

	return []string{
		w.Origin,
		w.Translation,
		w.LastSeenAt.Format(time.RFC3339),
		strconv.Itoa(w.Score),
		jsonCard,
	}
}

// GetVerbFormenCard requests Word card from verbformen site
func (w Word) GetVerbFormenCard() (*VerbFormenCard, error) {
	// todo:
	// 	This should be some universal method of MetaProvider

	ws := strings.Split(w.Origin, " ")
	r, err := http.Get(verbFormenURL + strings.Join(ws, "+"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Body.Close() }()
	node, err := html.Parse(r.Body)
	if err != nil {
		return nil, err
	}

	// i just need first <section> in page
	section := findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "section"
	})

	if section != nil {
		return NewVerbFormenCardFromNode(section), nil
	}

	return nil, errors.New("can't get Card")
}

func findNodes(n *html.Node, pred func(node *html.Node) bool) []*html.Node {
	var ns []*html.Node

	if pred(n) {
		ns = append(ns, n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		resN := findNodes(c, pred)
		if resN != nil {
			ns = append(ns, resN...)
		}
	}

	return ns
}

func findNode(n *html.Node, pred func(node *html.Node) bool) *html.Node {
	if pred(n) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		resN := findNode(c, pred)
		if resN != nil {
			return resN
		}
	}
	return nil
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

// SaveToCSV saves entire Words heap on disk like a CSV file
func (ws *Words) SaveToCSV(path string) (int, error) {
	f, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return -1, err
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(f)
	w.Comma = ';'
	// hint: schema
	titles := []string{"origin", "translation", "last_seen_at", "score", "meta"}
	err = w.Write(titles)
	if err != nil {
		return -1, err
	}
	for _, word := range *ws {
		err = w.Write(word.GetCSV())
		if err != nil {
			return -1, err
		}
	}
	w.Flush()
	return ws.Len(), nil
}

// LoadFromCSV loads words from CSV file
func (ws *Words) LoadFromCSV(path string) (int, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return -1, err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.Comma = ';'

	data, err := r.ReadAll()
	if err != nil {
		return -1, err
	}

	for _, row := range data[1:] {
		w, err := NewWordFromCSV(row)
		if err != nil {
			fmt.Printf("can't make a word from data: %v", err)
			continue
		}
		heap.Push(ws, w)
	}

	return len(data) - 1, nil
}
