package datastore

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/matt-FFFFFF/bookdata-api/loader"
)

// Books is the memory-backed datastore used by the API
// It contains a single field 'Store', which is (a pointer to) a slice of loader.BookData struct pointers
type Books struct {
	Store *[]*loader.BookData `json:"store"`
}

// Initialize is the method used to populate the in-memory datastore.
// At the beginning, this simply returns a pointer to the struct literal.
// You need to change this to load data from the CSV file
func (b *Books) Initialize() {
	//b.Store = &loader.BooksLiteral
	books := []*loader.BookData{}

	file, err := os.Open("assets/books.csv")
	if err != nil {
		log.Fatal("Unable to load books data file.")
	}
	defer file.Close()

	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		data := &loader.BookData{
			BookID:  record[0],
			Title:   record[1],
			Authors: record[2],
			ISBN:    record[4],
			ISBN13:  record[5],
		}
		if avgRating, err := strconv.ParseFloat(record[3], 32); err != nil {
			data.AverageRating = avgRating
		}
		books = append(books, data)
	}
	b.Store = &books
}

// GetAllBooks returns the entire dataset, subjet to the rudimentary limit & skip parameters
func (b *Books) GetAllBooks(limit, skip int) *[]*loader.BookData {
	if limit == 0 || limit > len(*b.Store) {
		limit = len(*b.Store)
	}
	ret := (*b.Store)[skip:limit]
	return &ret
}
