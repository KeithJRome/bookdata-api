package datastore

import "github.com/matt-FFFFFF/bookdata-api/loader"

// BookStore is the interface that the http methods use to call the backend datastore
// Using an interface means we could replace the datastore with something else,
// as long as that something else provides these method signatures...
type BookStore interface {
	Initialize()
	GetAllBooks(limit, skip int) *[]*loader.BookData
	SearchByAuthor(author string) *[]*loader.BookData
	SearchByTitle(title string) *[]*loader.BookData
	SearchByIsbn(isbn string) *[]*loader.BookData
	DeleteByIsbn(isbn string) bool
}
