package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/matt-FFFFFF/bookdata-api/datastore"
	"github.com/matt-FFFFFF/bookdata-api/loader"
)

func getLimitParam(r *http.Request) (limit int, err error) {
	queryParams := r.URL.Query()
	l := queryParams.Get("limit")
	if l != "" {
		var val int
		val, err = strconv.Atoi(l)
		if err != nil {
			err = fmt.Errorf("limit must be an integer - %w", err)
			return
		}
		if val < 0 {
			err = errors.New(("limit must be >= 0"))
			return
		}
		limit = val
	}
	return
}

func getSkipParam(r *http.Request) (skip int, err error) {
	queryParams := r.URL.Query()
	l := queryParams.Get("skip")
	if l != "" {
		var val int
		val, err = strconv.Atoi(l)
		if err != nil {
			err = fmt.Errorf("skip must be an integer - %w", err)
		}
		if val < 0 {
			err = errors.New("skip must be >= 0")
			return
		}
		skip = val
	}
	return
}

func returnBooks(w http.ResponseWriter, r *http.Request, getter func(limit, skip int) *[]*loader.BookData) {
	w.Header().Set("Content-Type", "application/json")

	// limit
	limit, err := getLimitParam(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := fmt.Sprintf(`{"error": %v}`, err)
		w.Write([]byte(resp))
		return
	}

	// skip
	skip, err := getSkipParam(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := fmt.Sprintf(`{"error": %v}`, err)
		w.Write([]byte(resp))
		return
	}

	// get books using the provided func
	list := getter(limit, skip)

	// convert to JSON
	b, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshalling data"}`))
		return
	}

	// respond
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return

}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	returnBooks(w, r, func(limit, skip int) *[]*loader.BookData {
		return books.GetAllBooks(limit, skip)
	})
}

func searchBooksByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if fragment, ok := vars["author"]; ok {
		returnBooks(w, r, func(limit, skip int) *[]*loader.BookData {
			return books.SearchByAuthor(fragment, limit, skip)
		})
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "an author (fragment) was not provided for the search"}`))
	}
}

func searchBooksByTitle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if fragment, ok := vars["title"]; ok {
		returnBooks(w, r, func(limit, skip int) *[]*loader.BookData {
			return books.SearchByTitle(fragment, limit, skip)
		})
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "a title (fragment) was not provided for the search"}`))
	}
}

func getBookByIsbn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get ISBN
	vars := mux.Vars(r)
	isbn, ok := vars["isbn"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "an ISBN was not provided"}`))
		return
	}

	// find the book
	book := books.GetByISBN(isbn)
	if book == nil {
		w.WriteHeader(http.StatusNotFound)
		resp := fmt.Sprintf(`{"error": "no book was found with ISBN %v"}`, isbn)
		w.Write([]byte(resp))
		return
	}

	// convert to JSON
	b, err := json.Marshal(book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshalling data"}`))
		return
	}

	// respond
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return

}

func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// parse the form
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "unable to parse form"}`))
		return
	}

	// create validation funcs
	applyString := func(field *string, key string, allowEmpty bool) bool {
		val := r.FormValue(key)
		if allowEmpty || len(val) > 0 {
			*field = val
			return true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			resp := fmt.Sprintf(`{"error": "you must supply %v"}`, key)
			w.Write([]byte(resp))
			return false
		}
	}
	applyFloat := func(field *float64, key string, allowEmpty bool) bool {
		str := r.FormValue(key)
		if allowEmpty && len(str) < 1 {
			*field = 0
			return true
		} else if val, err := strconv.ParseFloat(str, 64); err == nil {
			*field = val
			return true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			resp := fmt.Sprintf(`{"error": "you must supply %v as a floating-point value"}`, key)
			w.Write([]byte(resp))
			return false
		}
	}
	applyInt := func(field *int, key string, allowEmpty bool) bool {
		str := r.FormValue(key)
		if allowEmpty && len(str) < 1 {
			*field = 0
			return true
		} else if val, err := strconv.Atoi(str); err == nil {
			*field = val
			return true
		} else {
			w.WriteHeader(http.StatusBadRequest)
			resp := fmt.Sprintf(`{"error": "you must supply %v as a integer value"}`, key)
			w.Write([]byte(resp))
			return false
		}
	}

	// create a new book struct
	var book loader.BookData

	// apply all values or stop on error
	if applyString(&book.BookID, "book_id", false) &&
		applyString(&book.Title, "title", false) &&
		applyString(&book.Authors, "authors", false) &&
		applyFloat(&book.AverageRating, "average_rating", true) &&
		applyString(&book.ISBN, "isbn", false) &&
		applyString(&book.ISBN13, "isbn_13", false) &&
		applyString(&book.LanguageCode, "language_code", false) &&
		applyInt(&book.NumPages, "num_pages", false) &&
		applyInt(&book.Ratings, "ratings", true) &&
		applyInt(&book.Reviews, "reviews", true) {

		// attempt insert
		err := books.InsertBook(&book)
		var conflict *datastore.InsertBookConflictError
		if errors.As(err, &conflict) {
			w.WriteHeader(http.StatusConflict)
			resp := fmt.Sprintf(`{"error": "%v"}`, err.Error())
			w.Write([]byte(resp))
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "unknown error inserting the book"}`))
			return
		}

		// return the book if it was created
		b, err := json.Marshal(book)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "error marshalling data"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(b)

	}

}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get the ISBN
	vars := mux.Vars(r)
	isbn, ok := vars["isbn"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "an ISBN was not provided"}`))
		return
	}

	// delete the book
	book, err := books.DeleteBook(isbn)
	var notFound *datastore.DeleteBookNotFoundError
	if errors.As(err, &notFound) {
		w.WriteHeader(http.StatusNotFound)
		resp := fmt.Sprintf(`{"error": "%v"}`, err.Error())
		w.Write([]byte(resp))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "unknown error deleting the book"}`))
		return
	}

	// convert to JSON
	b, err := json.Marshal(book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "error marshalling data"}`))
		return
	}

	// respond
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return

}
