package provider

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/egregors/karten/pkg/store"
	"github.com/muesli/termenv"
	"golang.org/x/net/html"
)

//nolint:revive // it's just colors
const (
	Default int = iota
	Green
	Blue
)

var (
	color      = termenv.EnvColorProfile().Color
	greenStyle = termenv.Style{}.Foreground(color("46")).Styled
	blueStyle  = termenv.Style{}.Foreground(color("69")).Styled
)

// Syllable is the part words with a specific style (color).
// I want it to look like on VerbFormen (with word parts highlighting).
type Syllable struct {
	Val   string `json:"val"`
	Color int    `json:"color"`
}

// Card is representation of word card from VerbFormen
type Card struct {
	Origin      []string
	Translation []string
	Forms       []*Syllable
}

func (c Card) toStyledString() string {
	s := strings.Join(c.Origin, " ") + "\n"

	for _, w := range c.Forms {
		sub := w.Val
		switch w.Color {
		case Green:
			s += greenStyle(sub)
		case Blue:
			s += blueStyle(sub)
		default:
			s += sub
		}

	}
	return s
}

// IsEmpty returns false is Card is empty
func (c Card) IsEmpty() bool {
	if len(c.Origin) == 0 || len(c.Translation) == 0 || len(c.Forms) == 0 {
		return true
	}
	return false
}

// VerbFormen – remove service to get translations and forms: https://www.verbformen.com/
type VerbFormen struct {
	URL string
}

// GetMeta hits remove service to get Metadata for particular store.Word
func (v VerbFormen) GetMeta(w *store.Word) error {
	card, err := v.getCard(w.Origin)
	if err != nil {
		return err
	}

	w.Origin = strings.Join(card.Origin, " ")
	w.Translation = strings.Join(card.Translation, " ")
	w.Meta = card.toStyledString()

	return nil
}

// GetVerbFormenCard requests Word card from verbformen site
func (v VerbFormen) getCard(s string) (*Card, error) {
	ws := strings.Split(s, " ")
	r, err := http.Get(v.URL + strings.Join(ws, "+"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Body.Close() }()
	node, err := html.Parse(r.Body)
	if err != nil {
		return nil, err
	}

	// just need first <section> in page
	section := findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "section"
	})

	card := &Card{}
	if section != nil {
		parseNode(section, card)
	}

	if card.IsEmpty() {
		return nil, fmt.Errorf("can't find the word: %s", s)
	}

	return card, nil
}

// NewVerbFormenCardFromNode takes HTML node with `section` from verbformen.com and
// parses it into Card struct
// todo: too complex, need refactoring
func parseNode(node *html.Node, card *Card) {
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

	var fs []*Syllable
	// todo: switch
	for _, n := range textNodes {
		d := strings.ReplaceAll(n.Data, "\n", " ")
		if d == "" || d == " " {
			continue
		}

		// neutral
		if n.Parent.Data == "p" || n.Parent.Data == "b" {
			fs = append(fs, &Syllable{d, Default})
		}

		// green
		if n.Parent.Data == "i" {
			fs = append(fs, &Syllable{d, Green})
		}

		// blue
		if n.Parent.Data == "u" {
			fs = append(fs, &Syllable{d, Blue})
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

	card.Origin = o
	card.Translation = t
	card.Forms = fs
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
