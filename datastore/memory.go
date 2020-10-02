package datastore

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
	// record the duration of this operation
	defer func(t time.Time) {
		log.Printf("Initialize() completed in %v ms\n", time.Since(t).Milliseconds())
	}(time.Now())

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
			BookID:       record[0],
			Title:        record[1],
			Authors:      record[2],
			ISBN:         record[4],
			ISBN13:       record[5],
			LanguageCode: record[6],
		}
		if avgRating, err := strconv.ParseFloat(record[3], 32); err == nil {
			data.AverageRating = avgRating
		}
		if numPages, err := strconv.ParseInt(record[7], 10, 32); err == nil {
			data.NumPages = int(numPages)
		}
		if ratings, err := strconv.ParseInt(record[8], 10, 32); err == nil {
			data.Ratings = int(ratings)
		}
		if reviews, err := strconv.ParseInt(record[9], 10, 32); err == nil {
			data.Reviews = int(reviews)
		}
		books = append(books, data)
	}
	b.Store = &books
}

// GetAllBooks returns the entire dataset, subject to the rudimentary limit & skip parameters
func (b *Books) GetAllBooks(limit, skip int) *[]*loader.BookData {
	if limit == 0 || limit > len(*b.Store) {
		limit = len(*b.Store)
	}
	ret := (*b.Store)[skip:limit]
	return &ret
}

// SearchByAuthor returns the entire dataset, filtered by author (case-insensitive, partial matching)
func (b *Books) SearchByAuthor(author string) *[]*loader.BookData {
	author = strings.ToLower(author)
	results := make([]*loader.BookData, 0)
	for _, book := range *b.Store {
		if strings.Contains(strings.ToLower(book.Authors), author) {
			results = append(results, book)
		}
	}
	return &results
}
