package datastore

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matt-FFFFFF/bookdata-api/loader"
)

// Books is the memory-backed datastore used by the API
// It contains a single field 'Store', which is (a pointer to) a slice of loader.BookData struct pointers
type Books struct {
	Store *[]*loader.BookData `json:"store"`
	mutex sync.RWMutex
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
		panic(fmt.Errorf("couldn't open the file - %w", err))
	}
	defer file.Close()

	// read the CSV data
	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		panic(fmt.Errorf("couldn't parse the CSV data - %w", err))
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
		book.LanguageCode = line[6]
		if pages, err := strconv.Atoi(line[7]); err == nil {
			book.NumPages = pages
		}
		if ratings, err := strconv.Atoi(line[8]); err == nil {
			book.Ratings = ratings
		}
		if reviews, err := strconv.Atoi(line[9]); err == nil {
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

func applyLimitAndSkipToBooks(list *[]*loader.BookData, limit, skip int) *[]*loader.BookData {
	if skip > len(*list) {
		empty := make([]*loader.BookData, 0)
		return &empty
	}
	max := len(*list) - skip
	if limit == 0 || limit > max {
		limit = max
	}
	filter := (*list)[skip : skip+limit]
	return &filter
}

func (b *Books) filter(consider func(book *loader.BookData) (include bool)) *[]*loader.BookData {
	filtered := make([]*loader.BookData, 0)
	for _, book := range *b.Store {
		if consider(book) {
			filtered = append(filtered, book)
		}
	}
	return &filtered
}

// GetAllBooks returns the entire dataset, subjet to the rudimentary limit & skip parameters
func (b *Books) GetAllBooks(limit, skip int) *[]*loader.BookData {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// limit and skip on everything
	list := applyLimitAndSkipToBooks(b.Store, limit, skip)

	return list
}

func (b *Books) SearchByAuthor(fragment string, limit, skip int) *[]*loader.BookData {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// filter to the appropriate set
	fragment = strings.ToLower(fragment)
	filtered := b.filter(func(book *loader.BookData) bool {
		authors := strings.ToLower(book.Authors)
		return strings.Contains(authors, fragment)
	})

	// limit and skip
	list := applyLimitAndSkipToBooks(filtered, limit, skip)

	return list
}

func (b *Books) SearchByTitle(fragment string, limit, skip int) *[]*loader.BookData {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// filter to the appropriate set
	fragment = strings.ToLower(fragment)
	filtered := b.filter(func(book *loader.BookData) bool {
		title := strings.ToLower(book.Title)
		return strings.Contains(title, fragment)
	})

	// limit and skip
	list := applyLimitAndSkipToBooks(filtered, limit, skip)

	return list
}

func (b *Books) GetByISBN(isbn string) *loader.BookData {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// look for isbn match
	for _, book := range *b.Store {
		if book.ISBN == isbn {
			return book
		}
	}

	return nil
}

type InsertBookConflictError struct {
	Msg string
}

func (e *InsertBookConflictError) Error() string {
	return e.Msg
}

func (b *Books) InsertBook(book *loader.BookData) error {

	// NOTE: Originally I had started with a READ lock for checking conflicts and then a WRITE lock for the append.
	//  However, if there was ever a method that allowed for the ISBN to be changed, then we could have a race
	//  condition, so I decided to instead make the whole thing a WRITE lock.

	// write lock
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// check for any conflicts
	for _, existing := range *b.Store {
		if book.ISBN == existing.ISBN {
			return &InsertBookConflictError{fmt.Sprintf("InsertBook() ISBN conflict on %v", book.ISBN)}
		}
		if book.ISBN13 == existing.ISBN13 {
			return &InsertBookConflictError{fmt.Sprintf("InsertBook() ISBN13 conflict on %v", book.ISBN13)}
		}
	}

	// add to the slice
	*b.Store = append(*b.Store, book)

	return nil
}

type DeleteBookNotFoundError struct {
	Msg string
}

func (e *DeleteBookNotFoundError) Error() string {
	return e.Msg
}

func (b *Books) DeleteBook(isbn string) (book *loader.BookData, err error) {

	// NOTE: Originally I had started with a READ lock for finding the index position and then a WRITE lock for
	//  the delete. However, if there was ever a method that allowed for other changes of the store, there could
	//  be a race condition, so I decided to instead make the whole thing a WRITE lock.

	// write lock
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// find the book
	for i, existing := range *b.Store {
		if isbn == existing.ISBN {

			// delete the entry
			// NOTE: if there was a lot of data consider (a) linked list, (b) order being preserved, or (c) mark as deleted
			*b.Store = append((*b.Store)[:i], (*b.Store)[i+1:]...)

			// NOTE: I just wanted to show labeled outputs
			book = existing
			return

		}
	}

	// raise not found
	err = &DeleteBookNotFoundError{fmt.Sprintf("DeleteBook() ISBN %v not found", isbn)}
	return

}
