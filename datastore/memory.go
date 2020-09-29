package datastore

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
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

	// NOTE: from literal
	//b.Store = &loader.BooksLiteral

	// time how long it takes
	start := time.Now()

	// open the file
	file, err := os.Open("./assets/books.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// read the CSV data
	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		panic(err)
	}

	// allocate
	all := make([]*loader.BookData, len(lines))

	// convert data into struct
	for i, line := range lines {
		var book loader.BookData
		book.BookID = line[0]
		book.Title = line[1]
		book.Authors = line[2]
		if rating, err := strconv.ParseFloat(line[3], 64); err == nil {
			book.AverageRating = rating
		}
		book.ISBN = line[4]
		book.ISBN13 = line[5]
		book.LanguageCode = line[5]
		if pages, err := strconv.Atoi(line[6]); err == nil {
			book.NumPages = pages
		}
		if ratings, err := strconv.Atoi(line[7]); err == nil {
			book.Ratings = ratings
		}
		if reviews, err := strconv.Atoi(line[8]); err == nil {
			book.Reviews = reviews
		}
		all[i] = &book
	}

	// assign
	b.Store = &all

	// record how long it took
	elapsed := time.Since(start)
	log.Printf("loading the books took %v.\n", elapsed)

}

// GetAllBooks returns the entire dataset, subjet to the rudimentary limit & skip parameters
func (b *Books) GetAllBooks(limit, skip int) *[]*loader.BookData {
	if skip > len(*b.Store) {
		empty := make([]*loader.BookData, 0)
		return &empty
	}
	max := len(*b.Store) - skip
	if limit == 0 || limit > max {
		limit = max
	}
	ret := (*b.Store)[skip : skip+limit]
	return &ret
}
