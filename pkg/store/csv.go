package store

import (
	"container/heap"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	csvOrigin int = iota
	csvTranslation
	csvLastSeenAt
	csvScore
	csvMeta
)

// CSV is .csv store backend for words. Compliantly simple. Read full file from disk.
// Save method will override whole file.
type CSV struct {
	Path string
}

// NewCSV open creates new CSV store. Creates a new CSV file, if it does not exist.
func NewCSV(path string) (*CSV, error) {
	err := getPath(path)
	if err != nil {
		return nil, fmt.Errorf("can't make CSV file: %w", err)
	}
	return &CSV{Path: path}, nil
}

func getPath(path string) error {
	if isFileExist(path) {
		return nil
	}

	dir, _ := filepath.Split(path)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	err = createFile(path)
	if err != nil {
		return err
	}

	return nil
}

// loadAll loads all word from CSV file in random order
func (c CSV) loadAll() (ws Words, err error) {
	f, err := os.Open(filepath.Clean(c.Path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.Comma = ';'

	data, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range data[1:] {
		t, err := time.Parse(time.RFC3339, row[csvLastSeenAt])
		if err != nil {
			t = time.Time{}
		}

		score, err := strconv.Atoi(row[csvScore])
		if err != nil {
			score = 0
		}

		w := &Word{
			Origin:      row[csvOrigin],
			Translation: row[csvTranslation],
			LastSeenAt:  t,
			Score:       score,
			Meta:        row[csvMeta],
		}

		if err != nil {
			fmt.Printf("can't make a word from data: %v", err)
			continue
		}
		ws = append(ws, w)
	}
	return ws, nil
}

// saveAll saves all words into CSV file
func (c CSV) saveAll(ws Words) error {
	f, err := os.OpenFile(filepath.Clean(c.Path), os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	w := csv.NewWriter(f)
	w.Comma = ';'
	// hint: schema
	titles := []string{"origin", "translation", "last_seen_at", "score", "meta"}
	err = w.Write(titles)
	if err != nil {
		return err
	}
	for _, word := range ws {
		err = w.Write(toRow(*word))
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

// AddWord adds new word in words collection and saves on disc
func (c CSV) AddWord(w *Word) error {
	// todo: reduplication
	ws, err := c.loadAll()
	if err != nil {
		return err
	}
	ws = append(ws, w)
	err = c.saveAll(ws)
	if err != nil {
		return err
	}
	return nil
}

// GetWords loads words and put in into a heap according the score
func (c CSV) GetWords(n int) (*Words, error) {
	ws, err := c.loadAll()
	if err != nil {
		return nil, err
	}

	cut := &Words{}
	for i := 0; i < n && i < len(ws); i++ {
		heap.Push(cut, ws[i])
	}

	return cut, nil
}

// Save saves Word into CSV file
func (c CSV) Save(w *Word) error {
	ws, err := c.loadAll()
	if err != nil {
		return err
	}
	for i, word := range ws {
		if word.Origin == w.Origin {
			ws[i] = w
			break
		}
	}
	err = c.saveAll(ws)
	if err != nil {
		return err
	}
	return nil
}

// toRow perform serialization from Word to CSV row.
//	 Schema:
//	 	origin 				:: string
//		translation 		:: string
//		last_seen_at 		:: string[time.RFC3339]
//		score 				:: int
//		meta 				:: string
func toRow(w Word) []string {
	return []string{
		w.Origin,
		w.Translation,
		w.LastSeenAt.Format(time.RFC3339),
		strconv.Itoa(w.Score),
		w.Meta,
	}
}

func isFileExist(path string) bool {
	if _, err := os.Stat(filepath.Clean(path)); os.IsNotExist(err) {
		return false
	}
	return true
}

func createFile(path string) error {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return CSV{Path: path}.saveAll(nil)
}
